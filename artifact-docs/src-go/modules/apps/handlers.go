package apps

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/httpjson"
)

// Store é o que a rota usa do banco — o Repository em produção, um falso nos testes.
type Store interface {
	Hub(ctx context.Context) (Hub, error)
}

// Handlers são as rotas /api/apps. Protegidas: passam pelo auth.RequireUser.
type Handlers struct {
	store  Store
	logger *slog.Logger
}

// NewHandlers monta as rotas do hub de apps.
func NewHandlers(store Store, logger *slog.Logger) *Handlers {
	return &Handlers{store: store, logger: logger}
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
