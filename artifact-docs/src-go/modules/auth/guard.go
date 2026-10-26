// Package auth é o login do hub do lado do servidor (docs/auth.md): o porteiro de toda rota
// /api protegida, as rotas /api/auth/* e a pergunta "o dono já criou a conta?".
//
// Quem confere a assinatura do token é o lib/infra/neonauth; aqui mora a REGRA — quem pode
// entrar e o que a tela recebe.
package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/httpjson"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/neonauth"
)

// TokenVerifier confere o JWT do Neon Auth — o neonauth.Verifier em produção, um falso nos
// testes.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (neonauth.User, error)
}

// Settings é o que o porteiro precisa saber.
type Settings struct {
	// Tokens é nil quando NEON_AUTH_JWKS_URL não foi configurada: aí toda rota protegida
	// responde 503, e não 401 — o problema é do servidor, não de quem tenta entrar.
	Tokens TokenVerifier
	// AllowedEmails é a lista de quem pode usar o hub, já em minúsculas.
	AllowedEmails []string
	// NeonAuthURL é o endereço do login, entregue à tela em /api/auth/config.
	NeonAuthURL string
}

type userKey struct{}

// RequireUser é o porteiro de toda rota /api que não é pública.
//
// A regra mora aqui, e não na tela: a tela só esconde o que a pessoa não pode ver; quem
// impede é o servidor.
func RequireUser(settings Settings, next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if settings.Tokens == nil {
			httpjson.Error(writer, http.StatusServiceUnavailable, "O login ainda não foi configurado no servidor.")
			return
		}

		token, ok := bearerToken(request)
		if !ok {
			httpjson.Error(writer, http.StatusUnauthorized, "Entre para continuar.")
			return
		}

		user, err := settings.Tokens.Verify(request.Context(), token)
		if err != nil {
			httpjson.Error(writer, http.StatusUnauthorized, "Sua sessão expirou. Entre de novo.")
			return
		}

		if !isAllowed(settings.AllowedEmails, user.Email) {
			httpjson.Error(writer, http.StatusForbidden, "Esta conta não tem acesso ao hub.")
			return
		}

		next(writer, request.WithContext(context.WithValue(request.Context(), userKey{}, user)))
	}
}

// CurrentUser é quem o porteiro deixou passar. Só existe dentro de rota com RequireUser.
func CurrentUser(request *http.Request) neonauth.User {
	user, _ := request.Context().Value(userKey{}).(neonauth.User)

	return user
}

// WithUser põe o usuário na requisição como o porteiro poria. Existe para os testes dos
// outros módulos, que testam a rota sem passar pela criptografia do token.
func WithUser(request *http.Request, user neonauth.User) *http.Request {
	return request.WithContext(context.WithValue(request.Context(), userKey{}, user))
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
