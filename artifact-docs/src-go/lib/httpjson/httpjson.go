// Package httpjson é o jeito único de a API responder: JSON no corpo e erro sempre no mesmo
// formato, o que a SPA lê em svelte/src/lib/api/api-error.type.ts.
package httpjson

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorBody é o formato único de erro da API: { "message": "texto em português" }.
type ErrorBody struct {
	Message string `json:"message"`
}

// Write responde status com body em JSON.
func Write(writer http.ResponseWriter, status int, body any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(body)
}

// Error responde status com a mensagem que a tela vai mostrar.
func Error(writer http.ResponseWriter, status int, message string) {
	Write(writer, status, ErrorBody{Message: message})
}

// Fail registra a falha interna e responde 500 com uma frase que diz o que fazer. O motivo
// técnico fica no log; para a tela ele não ajuda e pode contar demais.
func Fail(logger *slog.Logger, writer http.ResponseWriter, request *http.Request, err error, message string) {
	logger.Error("request failed", "method", request.Method, "path", request.URL.Path, "error", err)
	Error(writer, http.StatusInternalServerError, message)
}
