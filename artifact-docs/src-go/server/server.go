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
func New(logger *slog.Logger, spa fs.FS) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", health)
	// Endereço de API que não existe é 404 em JSON — nunca o index.html da SPA, que chegaria
	// à tela como "resposta que não é JSON" e esconderia o erro de verdade.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "Endereço da API não encontrado.")
	})
	mux.Handle("/", spaHandler(spa))

	return logRequests(logger, mux)
}

// health responde se o processo está de pé. É o que o Playwright e o deploy esperam antes
// de seguir.
func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// errorBody é o formato único de erro da API — o mesmo que a SPA lê em
// svelte/src/lib/api/api-error.type.ts.
type errorBody struct {
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorBody{Message: message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		logger.Debug("request", "method", r.Method, "path", r.URL.Path)
	})
}
