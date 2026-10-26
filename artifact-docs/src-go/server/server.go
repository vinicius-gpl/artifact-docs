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

// New devolve o handler raiz do servidor.
func New(logger *slog.Logger, spa fs.FS, authCfg Auth) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", health)
	mux.HandleFunc("GET /api/auth/config", authConfig(authCfg))
	mux.HandleFunc("GET /api/auth/me", requireUser(authCfg, authMe))
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
