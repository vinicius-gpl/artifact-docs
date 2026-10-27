package apps

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/httpjson"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/database"
)

// maxBodyBytes é o maior corpo de cadastro aceito: três campos curtos.
const maxBodyBytes = 8 << 10

const notFoundMessage = "App não encontrado. Ele pode ter sido apagado."

// Store é o que as rotas usam do banco — o Repository em produção, um falso nos testes.
type Store interface {
	Hub(ctx context.Context) (Hub, error)
	Create(ctx context.Context, input Input) (Link, error)
	Update(ctx context.Context, id string, input Input) (Link, error)
	Delete(ctx context.Context, id string) error
}

// Checker confere um app logo depois de cadastrado ou editado — o Monitor em produção.
type Checker interface {
	CheckSoon(target Target)
}

// Handlers são as rotas /api/apps. Protegidas: passam pelo auth.RequireUser.
type Handlers struct {
	store   Store
	checker Checker
	logger  *slog.Logger
}

// NewHandlers monta as rotas do hub de apps. checker pode ser nil (sem monitor): o app fica
// "sem verificação" até a próxima rodada.
func NewHandlers(store Store, checker Checker, logger *slog.Logger) *Handlers {
	return &Handlers{store: store, checker: checker, logger: logger}
}

// Hub é GET /api/apps → appHubSchema: tipos e apps juntos, e a tela agrupa.
func (handlers *Handlers) Hub(writer http.ResponseWriter, request *http.Request) {
	hub, err := handlers.store.Hub(request.Context())
	if err != nil {
		httpjson.Fail(handlers.logger, writer, request, err, "Não consegui carregar os apps. Tente de novo em instantes.")
		return
	}

	httpjson.Write(writer, http.StatusOK, hub)
}

// Create é POST /api/apps, corpo appLinkInputSchema → appLinkSchema (201).
func (handlers *Handlers) Create(writer http.ResponseWriter, request *http.Request) {
	input, ok := readInput(writer, request)
	if !ok {
		return
	}

	link, err := handlers.store.Create(request.Context(), input)
	if err != nil {
		httpjson.Fail(handlers.logger, writer, request, err, "Não consegui cadastrar o app. Tente de novo em instantes.")
		return
	}

	handlers.checkSoon(link)
	httpjson.Write(writer, http.StatusCreated, link)
}

// Update é PATCH /api/apps/{id}, corpo appLinkInputSchema → appLinkSchema.
func (handlers *Handlers) Update(writer http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	if !idPattern.MatchString(id) {
		httpjson.Error(writer, http.StatusNotFound, notFoundMessage)
		return
	}
	input, ok := readInput(writer, request)
	if !ok {
		return
	}

	link, err := handlers.store.Update(request.Context(), id, input)
	if errors.Is(err, database.ErrNotFound) {
		httpjson.Error(writer, http.StatusNotFound, notFoundMessage)
		return
	}
	if err != nil {
		httpjson.Fail(handlers.logger, writer, request, err, "Não consegui salvar o app. Tente de novo em instantes.")
		return
	}

	handlers.checkSoon(link)
	httpjson.Write(writer, http.StatusOK, link)
}

// Delete é DELETE /api/apps/{id} → 204.
func (handlers *Handlers) Delete(writer http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	if !idPattern.MatchString(id) {
		httpjson.Error(writer, http.StatusNotFound, notFoundMessage)
		return
	}

	err := handlers.store.Delete(request.Context(), id)
	if errors.Is(err, database.ErrNotFound) {
		httpjson.Error(writer, http.StatusNotFound, notFoundMessage)
		return
	}
	if err != nil {
		httpjson.Fail(handlers.logger, writer, request, err, "Não consegui apagar o app. Tente de novo em instantes.")
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (handlers *Handlers) checkSoon(link Link) {
	if handlers.checker != nil {
		handlers.checker.CheckSoon(Target{ID: link.ID, URL: link.URL})
	}
}

// readInput lê e confere o corpo; quando ele é inválido, já respondeu 400.
func readInput(writer http.ResponseWriter, request *http.Request) (Input, bool) {
	var raw Input
	if err := json.NewDecoder(http.MaxBytesReader(writer, request.Body, maxBodyBytes)).Decode(&raw); err != nil {
		httpjson.Error(writer, http.StatusBadRequest, "Cadastro inválido.")
		return Input{}, false
	}

	input, problem := normalize(raw)
	if problem != "" {
		httpjson.Error(writer, http.StatusBadRequest, problem)
		return Input{}, false
	}

	return input, true
}
