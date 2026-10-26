// Package server monta as rotas HTTP do hub.
//
//	/api/...   a API (JSON)
//	/...       a SPA — arquivo que existe é servido; o resto cai no index.html, e quem
//	           decide a tela é o roteador do SvelteKit no navegador
package server

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
)

// Dependencies é tudo o que as rotas usam de fora. Store nil quer dizer "não configurado"
// (sem DATABASE_URL): o servidor sobe assim mesmo, e as rotas dele respondem 503.
type Dependencies struct {
	Auth      Auth
	Setup     SetupChecker
	Artifacts ArtifactStore
	Apps      AppStore
}

// routes guarda o que os handlers da API compartilham.
type routes struct {
	logger *slog.Logger
	deps   Dependencies
}

// New devolve o handler raiz do servidor. A lista de rotas da API é a mesma tabela ROUTES de
// svelte/src/lib/api/mock-server.ts — rota nova nasce lá e no contrato, e chega aqui depois.
func New(logger *slog.Logger, spa fs.FS, deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	api := &routes{logger: logger, deps: deps}
	guard := func(next http.HandlerFunc) http.HandlerFunc { return requireUser(deps.Auth, next) }

	mux.HandleFunc("GET /api/health", health)
	mux.HandleFunc("GET /api/auth/config", authConfig(deps.Auth))
	mux.HandleFunc("GET /api/auth/setup", api.authSetup)
	mux.HandleFunc("GET /api/auth/me", guard(authMe))
	mux.HandleFunc("GET /api/artifacts", guard(api.listArtifacts))
	mux.HandleFunc("GET /api/artifacts/{label}", guard(api.openArtifact))
	mux.HandleFunc("PATCH /api/artifacts/{label}", guard(api.patchArtifact))
	mux.HandleFunc("GET /api/apps", guard(api.appHub))
	// Endereço de API que não existe é 404 em JSON — nunca o index.html da SPA, que chegaria
	// à tela como "resposta que não é JSON" e esconderia o erro de verdade.
	mux.HandleFunc("/api/", func(writer http.ResponseWriter, _ *http.Request) {
		writeError(writer, http.StatusNotFound, "Endereço da API não encontrado.")
	})
	mux.Handle("/", spaHandler(spa))

	return logRequests(logger, mux)
}

// health responde se o processo está de pé. É o que o Playwright e o deploy esperam antes
// de seguir.
func health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

// errorBody é o formato único de erro da API — o mesmo que a SPA lê em
// svelte/src/lib/api/api-error.type.ts.
type errorBody struct {
	Message string `json:"message"`
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writeJSON(writer, status, errorBody{Message: message})
}

// fail registra a falha interna e responde 500 com uma frase que diz o que fazer. O motivo
// técnico fica no log; para a tela ele não ajuda e pode contar demais.
func (api *routes) fail(writer http.ResponseWriter, request *http.Request, err error, message string) {
	api.logger.Error("request failed", "method", request.Method, "path", request.URL.Path, "error", err)
	writeError(writer, http.StatusInternalServerError, message)
}

// databaseMissing responde 503 quando a rota precisa do banco e ele não foi configurado.
func databaseMissing(writer http.ResponseWriter) {
	writeError(writer, http.StatusServiceUnavailable, "O banco de dados ainda não foi configurado no servidor.")
}

func writeJSON(writer http.ResponseWriter, status int, body any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(body)
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(writer, request)
		logger.Debug("request", "method", request.Method, "path", request.URL.Path)
	})
}
