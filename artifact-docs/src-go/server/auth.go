package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/auth"
)

// TokenVerifier confere o JWT do Neon Auth — o auth.Verifier em produção, um falso nos testes.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (auth.User, error)
}

// Auth é o que o porteiro precisa saber.
type Auth struct {
	// Tokens é nil quando NEON_AUTH_JWKS_URL não foi configurada: aí toda rota protegida
	// responde 503, e não 401 — o problema é do servidor, não de quem tenta entrar.
	Tokens TokenVerifier
	// AllowedEmails é a lista de quem pode usar o hub, já em minúsculas.
	AllowedEmails []string
	// NeonAuthURL é o endereço do login, entregue à tela em /api/auth/config.
	NeonAuthURL string
}

type userKey struct{}

// requireUser é o porteiro de toda rota /api que não é pública (ver docs/auth.md).
//
// A regra mora aqui, e não na tela: a tela só esconde o que a pessoa não pode ver; quem
// impede é o servidor.
func requireUser(cfg Auth, next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if cfg.Tokens == nil {
			writeError(writer, http.StatusServiceUnavailable, "O login ainda não foi configurado no servidor.")
			return
		}

		token, ok := bearerToken(request)
		if !ok {
			writeError(writer, http.StatusUnauthorized, "Entre para continuar.")
			return
		}

		user, err := cfg.Tokens.Verify(request.Context(), token)
		if err != nil {
			writeError(writer, http.StatusUnauthorized, "Sua sessão expirou. Entre de novo.")
			return
		}

		if !isAllowed(cfg.AllowedEmails, user.Email) {
			writeError(writer, http.StatusForbidden, "Esta conta não tem acesso ao hub.")
			return
		}

		next(writer, request.WithContext(context.WithValue(request.Context(), userKey{}, user)))
	}
}

func bearerToken(request *http.Request) (string, bool) {
	scheme, token, found := strings.Cut(request.Header.Get("Authorization"), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", false
	}

	return token, true
}

func isAllowed(allowed []string, email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	for _, item := range allowed {
		if item == email {
			return true
		}
	}

	return false
}

// currentUser é quem o porteiro deixou passar. Só existe dentro de rota com requireUser.
func currentUser(request *http.Request) auth.User {
	user, _ := request.Context().Value(userKey{}).(auth.User)

	return user
}

// sessionUser é o sessionUserSchema de svelte/src/lib/contracts/user.schema.ts — mudou lá,
// muda aqui.
type sessionUser struct {
	ID    string  `json:"id"`
	Email string  `json:"email"`
	Name  string  `json:"name"`
	Image *string `json:"image"`
}

func authMe(writer http.ResponseWriter, request *http.Request) {
	user := currentUser(request)

	writeJSON(writer, http.StatusOK, sessionUser{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Image: user.Image,
	})
}

// authConfig é público: diz à tela ONDE o login acontece quando a URL não veio no build
// (ver ensureNeonAuthClient em svelte/src/lib/auth/neon-auth.client.ts). Não é segredo.
func authConfig(cfg Auth) http.HandlerFunc {
	return func(writer http.ResponseWriter, _ *http.Request) {
		writeJSON(writer, http.StatusOK, map[string]string{"neonAuthUrl": cfg.NeonAuthURL})
	}
}
