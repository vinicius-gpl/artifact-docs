// Package server compõe o hub: liga os módulos às rotas HTTP e cuida da vida do processo.
//
//	/api/...   a API (JSON), uma rota por handler dos módulos
//	/...       a SPA — arquivo que existe é servido; o resto cai no index.html, e quem
//	           decide a tela é o roteador do SvelteKit no navegador
package server

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/httpjson"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/web"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/modules/apps"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/modules/artifacts"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/modules/auth"
)

// Dependencies é tudo o que as rotas usam de fora. Store nil quer dizer "sem banco" (sem
// DATABASE_URL): o servidor sobe assim mesmo, e as rotas dele respondem 503.
type Dependencies struct {
	Auth      auth.Settings
	Owners    auth.OwnerChecker
	Artifacts artifacts.Store
	Apps      apps.Store
}

// New devolve o handler raiz do servidor. A lista de rotas da API é a mesma tabela ROUTES de
// svelte/src/lib/api/mock-server.ts — rota nova nasce lá e no contrato, e chega aqui depois.
func New(logger *slog.Logger, spa fs.FS, deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	guard := func(next http.HandlerFunc) http.HandlerFunc { return auth.RequireUser(deps.Auth, next) }

	login := auth.NewHandlers(deps.Auth, deps.Owners, logger)
	mux.HandleFunc("GET /api/health", health)
	mux.HandleFunc("GET /api/auth/config", login.Config)
	mux.HandleFunc("GET /api/auth/setup", needsDatabase(deps.Owners != nil, login.Setup))
	mux.HandleFunc("GET /api/auth/me", guard(login.Me))

	gallery := artifacts.NewHandlers(deps.Artifacts, logger)
	hasArtifacts := deps.Artifacts != nil
	mux.HandleFunc("GET /api/artifacts", guard(needsDatabase(hasArtifacts, gallery.List)))
	mux.HandleFunc("GET /api/artifacts/{label}", guard(needsDatabase(hasArtifacts, gallery.Open)))
	mux.HandleFunc("PATCH /api/artifacts/{label}", guard(needsDatabase(hasArtifacts, gallery.Update)))

	hub := apps.NewHandlers(deps.Apps, logger)
	mux.HandleFunc("GET /api/apps", guard(needsDatabase(deps.Apps != nil, hub.Hub)))

	// Endereço de API que não existe é 404 em JSON — nunca o index.html da SPA, que chegaria
	// à tela como "resposta que não é JSON" e esconderia o erro de verdade.
	mux.HandleFunc("/api/", func(writer http.ResponseWriter, _ *http.Request) {
		httpjson.Error(writer, http.StatusNotFound, "Endereço da API não encontrado.")
	})
	mux.Handle("/", web.Handler(spa))

	return logRequests(logger, mux)
}

// health responde se o processo está de pé. É o que o Playwright e o deploy esperam antes
// de seguir.
func health(writer http.ResponseWriter, _ *http.Request) {
	httpjson.Write(writer, http.StatusOK, map[string]string{"status": "ok"})
}

// needsDatabase troca a rota por um 503 quando o banco não foi configurado, em vez de deixar
// o handler cair num store nil.
func needsDatabase(available bool, next http.HandlerFunc) http.HandlerFunc {
	if available {
		return next
	}

	return func(writer http.ResponseWriter, _ *http.Request) {
		httpjson.Error(writer, http.StatusServiceUnavailable, "O banco de dados ainda não foi configurado no servidor.")
	}
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(writer, request)
		logger.Debug("request", "method", request.Method, "path", request.URL.Path)
	})
}
