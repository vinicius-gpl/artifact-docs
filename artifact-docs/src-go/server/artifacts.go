package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/store"
)

// maxPatchBytes é o maior corpo aceito num PATCH. O artifactPatchSchema tem três campos; um
// corpo maior que isto não é uma alteração, é engano ou abuso.
const maxPatchBytes = 4 << 10

// artifactKinds e artifactStatuses são os enums de svelte/src/lib/contracts/artifact.schema.ts
// (e os CHECK da migration 0001) — mudou lá, muda aqui.
var (
	artifactKinds    = map[string]bool{"document": true, "plan": true, "idea": true}
	artifactStatuses = map[string]bool{"draft": true, "active": true, "done": true, "archived": true}
)

// ArtifactStore é o que as rotas de artefato usam do banco — o store.Store em produção.
type ArtifactStore interface {
	ListArtifacts(ctx context.Context, ownerID string, filter store.ArtifactFilter) ([]store.Artifact, error)
	OpenArtifact(ctx context.Context, ownerID, label string) (store.ArtifactDetail, error)
	PatchArtifact(ctx context.Context, ownerID, label string, patch store.ArtifactPatch) (store.Artifact, error)
}

// listArtifacts é GET /api/artifacts?kind=&q= → artifactListSchema.
//
// `kind` fora do enum não filtra, em vez de dar erro — o mesmo que o mock faz. Um link velho
// com um kind que deixou de existir mostra tudo, em vez de uma tela de erro.
func (api *routes) listArtifacts(writer http.ResponseWriter, request *http.Request) {
	if api.deps.Artifacts == nil {
		databaseMissing(writer)
		return
	}

	query := request.URL.Query()
	filter := store.ArtifactFilter{Search: query.Get("q")}
	if artifactKinds[query.Get("kind")] {
		filter.Kind = query.Get("kind")
	}

	items, err := api.deps.Artifacts.ListArtifacts(request.Context(), currentUser(request).ID, filter)
	if err != nil {
		api.fail(writer, request, err, "Não consegui carregar os artefatos. Tente de novo em instantes.")
		return
	}

	writeJSON(writer, http.StatusOK, map[string][]store.Artifact{"items": items})
}

// openArtifact é GET /api/artifacts/{label} → artifactDetailSchema. Abrir conta como acesso.
func (api *routes) openArtifact(writer http.ResponseWriter, request *http.Request) {
	if api.deps.Artifacts == nil {
		databaseMissing(writer)
		return
	}

	detail, err := api.deps.Artifacts.OpenArtifact(request.Context(), currentUser(request).ID, request.PathValue("label"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(writer, http.StatusNotFound, "Artefato não encontrado. Ele pode ter sido apagado.")
		return
	}
	if err != nil {
		api.fail(writer, request, err, "Não consegui abrir o artefato. Tente de novo em instantes.")
		return
	}

	writeJSON(writer, http.StatusOK, detail)
}

// patchArtifact é PATCH /api/artifacts/{label}, corpo artifactPatchSchema → artifactSchema.
func (api *routes) patchArtifact(writer http.ResponseWriter, request *http.Request) {
	if api.deps.Artifacts == nil {
		databaseMissing(writer)
		return
	}

	patch, ok := readArtifactPatch(writer, request)
	if !ok {
		return
	}

	artifact, err := api.deps.Artifacts.PatchArtifact(request.Context(), currentUser(request).ID, request.PathValue("label"), patch)
	if errors.Is(err, store.ErrNotFound) {
		writeError(writer, http.StatusNotFound, "Artefato não encontrado. Ele pode ter sido apagado.")
		return
	}
	if err != nil {
		api.fail(writer, request, err, "Não consegui salvar a alteração. Tente de novo em instantes.")
		return
	}

	writeJSON(writer, http.StatusOK, artifact)
}

// artifactPatchBody é o corpo do PATCH. O status precisa distinguir "não veio" de "veio
// null" (limpar a situação) — por isso o optionalStatus, e não um *string simples.
type artifactPatchBody struct {
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

// readArtifactPatch lê e confere o corpo; quando ele é inválido, já respondeu 400.
func readArtifactPatch(writer http.ResponseWriter, request *http.Request) (store.ArtifactPatch, bool) {
	var body artifactPatchBody
	err := json.NewDecoder(http.MaxBytesReader(writer, request.Body, maxPatchBytes)).Decode(&body)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "Alteração inválida.")
		return store.ArtifactPatch{}, false
	}

	if body.Status.Value != nil && !artifactStatuses[*body.Status.Value] {
		writeError(writer, http.StatusBadRequest, "Situação inválida.")
		return store.ArtifactPatch{}, false
	}

	return store.ArtifactPatch{
		IsPinned:  body.IsPinned,
		IsPrivate: body.IsPrivate,
		StatusSet: body.Status.Set,
		Status:    body.Status.Value,
	}, true
}
