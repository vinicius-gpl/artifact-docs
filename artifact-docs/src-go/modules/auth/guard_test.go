package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/httpjson"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/neonauth"
)

// fakeTokens aceita só os tokens do mapa — o porteiro é testado sem criptografia, que já tem
// teste próprio em lib/infra/neonauth.
type fakeTokens map[string]neonauth.User

func (tokens fakeTokens) Verify(_ context.Context, token string) (neonauth.User, error) {
	user, ok := tokens[token]
	if !ok {
		return neonauth.User{}, neonauth.ErrInvalidToken
	}

	return user, nil
}

var ownerSettings = Settings{
	Tokens: fakeTokens{
		"owner":    {ID: "u1", Email: "Dono@Exemplo.com", Name: "Dono"},
		"stranger": {ID: "u2", Email: "outro@exemplo.com", Name: "Outro"},
	},
	AllowedEmails: []string{"dono@exemplo.com"},
	NeonAuthURL:   "https://auth.example.com",
}

// guarded passa a requisição pelo porteiro e devolve a resposta e quem chegou do outro lado.
func guarded(test *testing.T, settings Settings, token string) (*http.Response, neonauth.User) {
	test.Helper()

	var reached neonauth.User
	next := func(writer http.ResponseWriter, request *http.Request) {
		reached = CurrentUser(request)
		writer.WriteHeader(http.StatusNoContent)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/anything", nil)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	RequireUser(settings, next).ServeHTTP(recorder, request)

	return recorder.Result(), reached
}

// feliz
func TestTheGateLetsTheOwnerThroughWithTheUserInTheRequest(test *testing.T) {
	response, reached := guarded(test, ownerSettings, "owner")

	if response.StatusCode != http.StatusNoContent {
		test.Fatalf("status = %d, want the next handler (204)", response.StatusCode)
	}
	if reached.ID != "u1" {
		test.Fatalf("user = %+v, want u1 (the allowed list ignores letter case)", reached)
	}
}

// feliz
func TestTestModeLetsTheTestUserInWithoutToken(test *testing.T) {
	response, reached := guarded(test, Settings{TestMode: true}, "")

	if response.StatusCode != http.StatusNoContent {
		test.Fatalf("status = %d, want the next handler (204)", response.StatusCode)
	}
	if reached != TestUser {
		test.Fatalf("user = %+v, want the test user", reached)
	}
}

// triste
func TestTheGateKeepsOutWhoCannotProveOrIsNotAllowed(test *testing.T) {
	cases := []struct {
		name     string
		settings Settings
		token    string
		status   int
	}{
		{"no token", ownerSettings, "", http.StatusUnauthorized},
		{"invalid token", ownerSettings, "forged", http.StatusUnauthorized},
		{"valid but not allowed", ownerSettings, "stranger", http.StatusForbidden},
		{"login not configured", Settings{}, "owner", http.StatusServiceUnavailable},
	}

	for _, testCase := range cases {
		response, reached := guarded(test, testCase.settings, testCase.token)

		if response.StatusCode != testCase.status {
			test.Errorf("%s: status = %d, want %d", testCase.name, response.StatusCode, testCase.status)
		}
		if reached.ID != "" {
			test.Errorf("%s: the next handler must not run", testCase.name)
		}

		var body httpjson.ErrorBody
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil || body.Message == "" {
			test.Errorf("%s: want a JSON error with a message, got %v", testCase.name, err)
		}
	}
}
