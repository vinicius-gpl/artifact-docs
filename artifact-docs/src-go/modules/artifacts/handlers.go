package artifacts

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/httpjson"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/database"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/modules/auth"
)

// maxPatchBytes é o maior corpo aceito num PATCH. O artifactPatchSchema tem três campos; um
// corpo maior que isto não é uma alteração, é engano ou abuso.
const maxPatchBytes = 4 << 10

const notFoundMessage = "Artefato não encontrado. Ele pode ter sido apagado."

// Store é o que as rotas usam do banco — o Repository em produção, um falso nos testes.
type Store interface {
	List(ctx context.Context, ownerID string, filter Filter) ([]Artifact, error)
	Open(ctx context.Context, ownerID, label string) (Detail, error)
	Update(ctx context.Context, ownerID, label string, patch Patch) (Artifact, error)
}

// Handlers são as rotas /api/artifacts. Todas protegidas: passam pelo auth.RequireUser.
type Handlers struct {
	store  Store
	logger *slog.Logger
}

// NewHandlers monta as rotas de artefato.
func NewHandlers(store Store, logger *slog.Logger) *Handlers {
	return &Handlers{store: store, logger: logger}
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

	httpjson.Write(writer, http.StatusOK, map[string][]Artifact{"items": items})
}

// Open é GET /api/artifacts/{label} → artifactDetailSchema. Abrir conta como acesso.
func (handlers *Handlers) Open(writer http.ResponseWriter, request *http.Request) {
	detail, err := handlers.store.Open(request.Context(), auth.CurrentUser(request).ID, request.PathValue("label"))
	if errors.Is(err, database.ErrNotFound) {
		httpjson.Error(writer, http.StatusNotFound, notFoundMessage)
		return
	}
	if err != nil {
		httpjson.Fail(handlers.logger, writer, request, err, "Não consegui abrir o artefato. Tente de novo em instantes.")
		return
	}

	httpjson.Write(writer, http.StatusOK, detail)
}

// Update é PATCH /api/artifacts/{label}, corpo artifactPatchSchema → artifactSchema.
func (handlers *Handlers) Update(writer http.ResponseWriter, request *http.Request) {
	patch, ok := readPatch(writer, request)
	if !ok {
		return
	}

	artifact, err := handlers.store.Update(request.Context(), auth.CurrentUser(request).ID, request.PathValue("label"), patch)
	if errors.Is(err, database.ErrNotFound) {
		httpjson.Error(writer, http.StatusNotFound, notFoundMessage)
		return
	}
	if err != nil {
		httpjson.Fail(handlers.logger, writer, request, err, "Não consegui salvar a alteração. Tente de novo em instantes.")
		return
	}

	httpjson.Write(writer, http.StatusOK, artifact)
}

// patchBody é o corpo do PATCH. O status precisa distinguir "não veio" de "veio null"
// (limpar a situação) — por isso o optionalStatus, e não um *string simples.
type patchBody struct {
	IsPinned  *bool          `json:"isPinned"`
	IsPrivate *bool          `json:"isPrivate"`
	Status    optionalStatus `json:"status"`
}

type optionalStatus struct {
	Set   bool
	Value *string
}

// UnmarshalJSON só é chamado quando a chave está no corpo — inclusive com null. É isso que
// marca o "veio".
func (status *optionalStatus) UnmarshalJSON(raw []byte) error {
	status.Set = true

	return json.Unmarshal(raw, &status.Value)
}

// readPatch lê e confere o corpo; quando ele é inválido, já respondeu 400.
func readPatch(writer http.ResponseWriter, request *http.Request) (Patch, bool) {
	var body patchBody
	if err := json.NewDecoder(http.MaxBytesReader(writer, request.Body, maxPatchBytes)).Decode(&body); err != nil {
		httpjson.Error(writer, http.StatusBadRequest, "Alteração inválida.")
		return Patch{}, false
	}

	if body.Status.Value != nil && !knownStatuses[*body.Status.Value] {
		httpjson.Error(writer, http.StatusBadRequest, "Situação inválida.")
		return Patch{}, false
	}

	return Patch{
		IsPinned:  body.IsPinned,
		IsPrivate: body.IsPrivate,
		StatusSet: body.Status.Set,
		Status:    body.Status.Value,
	}, true
}
