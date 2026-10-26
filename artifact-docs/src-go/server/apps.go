package server

import (
	"context"
	"net/http"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/store"
)

// AppStore é o que a rota do hub de apps usa do banco — o store.Store em produção.
type AppStore interface {
	AppHub(ctx context.Context) (store.AppHub, error)
}

// appHub é GET /api/apps → appHubSchema: tipos e apps juntos, e a tela agrupa.
func (api *routes) appHub(writer http.ResponseWriter, request *http.Request) {
	if api.deps.Apps == nil {
		databaseMissing(writer)
		return
	}

	hub, err := api.deps.Apps.AppHub(request.Context())
	if err != nil {
		api.fail(writer, request, err, "Não consegui carregar os apps. Tente de novo em instantes.")
		return
	}

	writeJSON(writer, http.StatusOK, hub)
}
