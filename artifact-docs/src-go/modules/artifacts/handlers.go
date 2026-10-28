package artifacts

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/httpjson"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/database"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/storage"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/modules/auth"
)

// maxJSONBytes é o maior corpo JSON aceito (PATCH, restaurar). Cabe a descrição inteira (até
// maxFieldBytes) com título e tags; um corpo maior que isto não é uma alteração, é engano ou
// abuso.
const maxJSONBytes = 16 << 10

const notFoundMessage = "Artefato não encontrado. Ele pode ter sido apagado."

// Store é o que as rotas usam do banco — o Repository em produção, um falso nos testes.
type Store interface {
	List(ctx context.Context, ownerID string, filter Filter) ([]Artifact, error)
	Open(ctx context.Context, ownerID, label string) (Detail, error)
	Find(ctx context.Context, ownerID, label string) (Artifact, error)
	Update(ctx context.Context, ownerID, label string, patch Patch) (Artifact, error)
	FindServed(ctx context.Context, linkToken string) (Served, error)
	PublishTarget(ctx context.Context, ownerID, label string) (PublishTarget, error)
	SaveVersion(ctx context.Context, record VersionRecord) error
	Restore(ctx context.Context, ownerID, label string, version int) (string, error)
	Delete(ctx context.Context, ownerID, label string) (Deleted, error)
}

// Files é onde moram os arquivos das versões — o storage.Bucket (R2) em produção.
type Files interface {
	Put(ctx context.Context, key string, content io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (storage.Object, error)
	DeleteFolder(ctx context.Context, prefix string) error
}

// Handlers são as rotas /api/artifacts (protegidas: passam pelo auth.RequireUser) e o
// /a/<acesso>/ que serve os artefatos (público: quem decide é o link).
type Handlers struct {
	store  Store
	files  Files
	linker *Linker
	served *servedCache
	logger *slog.Logger
}

// NewHandlers monta as rotas de artefato. files é nil quando o R2 não foi configurado: aí
// só as rotas de dados funcionam (quem compõe o servidor responde 503 nas outras).
func NewHandlers(store Store, files Files, linker *Linker, logger *slog.Logger) *Handlers {
	return &Handlers{store: store, files: files, linker: linker, served: newServedCache(), logger: logger}
}

// List é GET /api/artifacts?kind=&q= → artifactListSchema.
//
// `kind` fora do enum não filtra, em vez de dar erro — o mesmo que o mock faz. Um link velho
// com um kind que deixou de existir mostra tudo, em vez de uma tela de erro.
func (handlers *Handlers) List(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	filter := Filter{Search: query.Get("q")}
	if knownKinds[query.Get("kind")] {
		filter.Kind = query.Get("kind")
	}

	items, err := handlers.store.List(request.Context(), auth.CurrentUser(request).ID, filter)
	if err != nil {
		httpjson.Fail(handlers.logger, writer, request, err, "Não consegui carregar os artefatos. Tente de novo em instantes.")
		return
	}

	for index := range items {
		items[index] = handlers.linker.withURLs(items[index])
	}
	httpjson.Write(writer, http.StatusOK, map[string][]Artifact{"items": items})
}

// Open é GET /api/artifacts/{label} → artifactDetailSchema. Abrir conta como acesso.
func (handlers *Handlers) Open(writer http.ResponseWriter, request *http.Request) {
	detail, err := handlers.store.Open(request.Context(), auth.CurrentUser(request).ID, request.PathValue("label"))
	if handlers.answeredFailure(writer, request, err, notFoundMessage, "Não consegui abrir o artefato.") {
		return
	}

	detail.Artifact = handlers.linker.withURLs(detail.Artifact)
	httpjson.Write(writer, http.StatusOK, detail)
}

// Update é PATCH /api/artifacts/{label}, corpo artifactPatchSchema → artifactSchema.
func (handlers *Handlers) Update(writer http.ResponseWriter, request *http.Request) {
	patch, ok := readPatch(writer, request)
	if !ok {
		return
	}

	artifact, err := handlers.store.Update(request.Context(), auth.CurrentUser(request).ID, request.PathValue("label"), patch)
	if handlers.answeredFailure(writer, request, err, notFoundMessage, "Não consegui salvar a alteração.") {
		return
	}

	// Virar privado precisa valer já: o cache não pode continuar servindo pelo link público.
	handlers.served.forget(artifact.linkToken)
	httpjson.Write(writer, http.StatusOK, handlers.linker.withURLs(artifact))
}

// Restore é POST /api/artifacts/{label}/restore, corpo { version } → artifactSchema: o
// rollback, que só troca o ponteiro da versão servida.
func (handlers *Handlers) Restore(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		Version int `json:"version"`
	}
	err := json.NewDecoder(http.MaxBytesReader(writer, request.Body, maxJSONBytes)).Decode(&body)
	if err != nil || body.Version < 1 {
		httpjson.Error(writer, http.StatusBadRequest, "Informe a versão a restaurar.")
		return
	}

	ownerID, label := auth.CurrentUser(request).ID, request.PathValue("label")
	linkToken, err := handlers.store.Restore(request.Context(), ownerID, label, body.Version)
	if handlers.answeredFailure(writer, request, err, "Versão não encontrada.", "Não consegui restaurar a versão.") {
		return
	}
	handlers.served.forget(linkToken)

	artifact, err := handlers.store.Find(request.Context(), ownerID, label)
	if handlers.answeredFailure(writer, request, err, notFoundMessage, "Não consegui carregar o artefato.") {
		return
	}
	httpjson.Write(writer, http.StatusOK, handlers.linker.withURLs(artifact))
}

// Delete é DELETE /api/artifacts/{label} → 204. Apaga o registro primeiro e os arquivos
// depois: se o R2 falhar, sobra arquivo sem dono (lixo que não aparece), e nunca um card
// apontando para arquivo que não existe.
func (handlers *Handlers) Delete(writer http.ResponseWriter, request *http.Request) {
	deleted, err := handlers.store.Delete(request.Context(), auth.CurrentUser(request).ID, request.PathValue("label"))
	if handlers.answeredFailure(writer, request, err, notFoundMessage, "Não consegui apagar o artefato.") {
		return
	}
	handlers.served.forget(deleted.LinkToken)

	if err := handlers.files.DeleteFolder(request.Context(), artifactFolder(deleted.ArtifactID)); err != nil {
		handlers.logger.Error("artifact files left behind", "artifact", deleted.ArtifactID, "error", err)
	}
	writer.WriteHeader(http.StatusNoContent)
}

// answeredFailure responde o erro (404 com notFound, 500 com failure) e diz se respondeu.
func (handlers *Handlers) answeredFailure(writer http.ResponseWriter, request *http.Request, err error, notFound, failure string) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, database.ErrNotFound) {
		httpjson.Error(writer, http.StatusNotFound, notFound)
		return true
	}

	httpjson.Fail(handlers.logger, writer, request, err, failure+" Tente de novo em instantes.")

	return true
}

// patchBody é o corpo do PATCH. Status e descrição precisam distinguir "não veio" de "veio
// null" (limpar) — por isso o optionalText, e não um *string simples. Tags `null` ou ausentes
// mantêm; `[]` limpa.
type patchBody struct {
	IsPinned    *bool        `json:"isPinned"`
	IsPrivate   *bool        `json:"isPrivate"`
	Status      optionalText `json:"status"`
	Title       *string      `json:"title"`
	Description optionalText `json:"description"`
	Tags        []string     `json:"tags"`
}

type optionalText struct {
	Set   bool
	Value *string
}

// UnmarshalJSON só é chamado quando a chave está no corpo — inclusive com null. É isso que
// marca o "veio".
func (text *optionalText) UnmarshalJSON(raw []byte) error {
	text.Set = true

	return json.Unmarshal(raw, &text.Value)
}

// readPatch lê e confere o corpo; quando ele é inválido, já respondeu 400.
func readPatch(writer http.ResponseWriter, request *http.Request) (Patch, bool) {
	var body patchBody
	if err := json.NewDecoder(http.MaxBytesReader(writer, request.Body, maxJSONBytes)).Decode(&body); err != nil {
		httpjson.Error(writer, http.StatusBadRequest, "Alteração inválida.")
		return Patch{}, false
	}

	patch, err := body.toPatch()
	if err != nil {
		httpjson.Error(writer, http.StatusBadRequest, err.Error())
		return Patch{}, false
	}

	return patch, true
}

// toPatch confere cada campo com a MESMA regra da publicação: editar não pode deixar entrar
// o que publicar recusaria.
func (body patchBody) toPatch() (Patch, error) {
	if body.Status.Value != nil && !knownStatuses[*body.Status.Value] {
		return Patch{}, invalidUpload{"Situação inválida."}
	}

	patch := Patch{
		IsPinned:       body.IsPinned,
		IsPrivate:      body.IsPrivate,
		StatusSet:      body.Status.Set,
		Status:         body.Status.Value,
		DescriptionSet: body.Description.Set,
	}

	if err := patch.setTitle(body.Title); err != nil {
		return Patch{}, err
	}
	if err := patch.setDescription(body.Description.Value); err != nil {
		return Patch{}, err
	}
	if body.Tags == nil {
		return patch, nil
	}

	tags, err := normalizeTags(body.Tags)
	patch.Tags = tags

	return patch, err
}

func (patch *Patch) setTitle(raw *string) error {
	if raw == nil {
		return nil
	}

	title := strings.TrimSpace(*raw)
	if err := checkTitle(title); err != nil {
		return err
	}
	patch.Title = &title

	return nil
}

// setDescription apara a descrição; vazia vira NULL no banco, como na publicação.
func (patch *Patch) setDescription(raw *string) error {
	if raw == nil {
		return nil
	}

	description := strings.TrimSpace(*raw)
	if len(description) > maxFieldBytes {
		return invalidUpload{"A descrição passa de 4 KB."}
	}
	patch.Description = &description

	return nil
}
