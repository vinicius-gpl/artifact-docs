package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/httpjson"
)

// OwnerChecker responde se o dono já criou a conta — o OwnerRepository em produção.
type OwnerChecker interface {
	HasOwner(ctx context.Context, allowedEmails []string) (bool, error)
}

// Handlers são as rotas /api/auth/*.
type Handlers struct {
	settings Settings
	owners   OwnerChecker
	logger   *slog.Logger
}

// NewHandlers monta as rotas de login.
func NewHandlers(settings Settings, owners OwnerChecker, logger *slog.Logger) *Handlers {
	return &Handlers{settings: settings, owners: owners, logger: logger}
}

// sessionUser é o sessionUserSchema de svelte/src/lib/contracts/user.schema.ts — mudou lá,
// muda aqui.
type sessionUser struct {
	ID    string  `json:"id"`
	Email string  `json:"email"`
	Name  string  `json:"name"`
	Image *string `json:"image"`
}

// Me é GET /api/auth/me → sessionUserSchema. Protegida: passa pelo RequireUser.
func (handlers *Handlers) Me(writer http.ResponseWriter, request *http.Request) {
	user := CurrentUser(request)

	httpjson.Write(writer, http.StatusOK, sessionUser{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Image: user.Image,
	})
}

// Config é GET /api/auth/config, pública: diz à tela ONDE o login acontece quando a URL não
// veio no build (ensureNeonAuthClient em svelte/src/lib/auth/neon-auth.client.ts). Não é
// segredo.
func (handlers *Handlers) Config(writer http.ResponseWriter, _ *http.Request) {
	httpjson.Write(writer, http.StatusOK, map[string]string{"neonAuthUrl": handlers.settings.NeonAuthURL})
}

// Setup é GET /api/auth/setup, pública: a tela de login pergunta, antes de qualquer sessão,
// se deve oferecer o "Primeiro acesso" (setupStatusSchema em
// svelte/src/lib/contracts/auth.schema.ts).
//
// No modo de teste a conta "já existe": é o usuário de teste, e o e2e entra direto.
func (handlers *Handlers) Setup(writer http.ResponseWriter, request *http.Request) {
	if handlers.settings.TestMode {
		httpjson.Write(writer, http.StatusOK, map[string]bool{"needsSetup": false})
		return
	}

	hasOwner, err := handlers.owners.HasOwner(request.Context(), handlers.settings.AllowedEmails)
	if err != nil {
		httpjson.Fail(handlers.logger, writer, request, err, "Não consegui conferir as contas. Tente de novo em instantes.")
		return
	}

	httpjson.Write(writer, http.StatusOK, map[string]bool{"needsSetup": !hasOwner})
}
