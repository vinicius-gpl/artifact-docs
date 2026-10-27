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
	// AppChecker confere um app logo depois de cadastrado (o monitor). Nil, ele espera a
	// próxima rodada.
	AppChecker apps.Checker
	// Files é o R2. Nil (sem R2_*), publicar, apagar e abrir artefato respondem 503.
	Files artifacts.Files
	// LinkSecret assina os links dos artefatos privados.
	LinkSecret []byte
}

// New devolve o handler raiz do servidor. A lista de rotas da API é a mesma tabela ROUTES de
// svelte/src/lib/api/mock-server.ts — rota nova nasce lá e no contrato, e chega aqui depois.
func New(logger *slog.Logger, spa fs.FS, deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	guard := func(next http.HandlerFunc) http.HandlerFunc { return auth.RequireUser(deps.Auth, next) }

	login := auth.NewHandlers(deps.Auth, deps.Owners, logger)
	mux.HandleFunc("GET /api/health", health)
	mux.HandleFunc("GET /api/auth/config", login.Config)
	mux.HandleFunc("GET /api/auth/setup", needsDatabase(deps.Owners != nil || deps.Auth.TestMode, login.Setup))
	mux.HandleFunc("GET /api/auth/me", guard(login.Me))

	gallery := artifacts.NewHandlers(deps.Artifacts, deps.Files, artifacts.NewLinker(deps.LinkSecret), logger)
	withData := func(next http.HandlerFunc) http.HandlerFunc { return needsDatabase(deps.Artifacts != nil, next) }
	withFiles := func(next http.HandlerFunc) http.HandlerFunc { return withData(needsStorage(deps.Files != nil, next)) }
	mux.HandleFunc("GET /api/artifacts", guard(withData(gallery.List)))
	mux.HandleFunc("GET /api/artifacts/{label}", guard(withData(gallery.Open)))
	mux.HandleFunc("PATCH /api/artifacts/{label}", guard(withData(gallery.Update)))
	mux.HandleFunc("PUT /api/artifacts/{label}", guard(withFiles(gallery.Publish)))
	mux.HandleFunc("POST /api/artifacts/{label}/restore", guard(withData(gallery.Restore)))
	mux.HandleFunc("DELETE /api/artifacts/{label}", guard(withFiles(gallery.Delete)))
	// Os artefatos servidos: públicos, porque quem decide é o link (links.go).
	mux.HandleFunc("GET /a/{access}", gallery.RedirectToFolder)
	mux.HandleFunc("GET /a/{access}/{path...}", withFiles(gallery.Serve))

	hub := apps.NewHandlers(deps.Apps, deps.AppChecker, logger)
	withApps := func(next http.HandlerFunc) http.HandlerFunc { return guard(needsDatabase(deps.Apps != nil, next)) }
	mux.HandleFunc("GET /api/apps", withApps(hub.Hub))
	mux.HandleFunc("POST /api/apps", withApps(hub.Create))
	mux.HandleFunc("PATCH /api/apps/{id}", withApps(hub.Update))
	mux.HandleFunc("DELETE /api/apps/{id}", withApps(hub.Delete))

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

// needsStorage é o needsDatabase do R2.
func needsStorage(available bool, next http.HandlerFunc) http.HandlerFunc {
	if available {
		return next
	}

	return func(writer http.ResponseWriter, _ *http.Request) {
		httpjson.Error(writer, http.StatusServiceUnavailable, "O armazenamento de arquivos (R2) ainda não foi configurado no servidor.")
	}
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(writer, request)
		logger.Debug("request", "method", request.Method, "path", request.URL.Path)
	})
}
