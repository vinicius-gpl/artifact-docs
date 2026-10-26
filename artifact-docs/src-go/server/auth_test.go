package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/auth"
)

// fakeTokens aceita só os tokens do mapa — o porteiro é testado sem criptografia, que já tem
// teste próprio em src-go/auth.
type fakeTokens map[string]auth.User

func (tokens fakeTokens) Verify(_ context.Context, token string) (auth.User, error) {
	user, ok := tokens[token]
	if !ok {
		return auth.User{}, auth.ErrInvalidToken
	}

	return user, nil
}

var allowedAuth = Auth{
	Tokens: fakeTokens{
		"owner":    {ID: "u1", Email: "Dono@Exemplo.com", Name: "Dono"},
		"stranger": {ID: "u2", Email: "outro@exemplo.com", Name: "Outro"},
	},
	AllowedEmails: []string{"dono@exemplo.com"},
	NeonAuthURL:   "https://auth.example.com",
}

func getWithToken(test *testing.T, cfg Auth, target, token string) *http.Response {
	test.Helper()

	req := httptest.NewRequest(http.MethodGet, target, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	New(slog.New(slog.NewTextHandler(io.Discard, nil)), builtSPA, cfg).ServeHTTP(rec, req)

	return rec.Result()
}

// feliz
func TestMeAnswersWhoIsLoggedIn(test *testing.T) {
	res := getWithToken(test, allowedAuth, "/api/auth/me", "owner")

	if res.StatusCode != http.StatusOK {
		test.Fatalf("status = %d, want 200", res.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		test.Fatalf("decode: %v", err)
	}
	if body["id"] != "u1" || body["email"] != "Dono@Exemplo.com" || body["name"] != "Dono" {
		test.Fatalf("body = %v", body)
	}
	if image, ok := body["image"]; !ok || image != nil {
		test.Fatalf("image = %v, want null (the contract requires the key)", image)
	}
}

// feliz
func TestAuthConfigIsPublic(test *testing.T) {
	res := getWithToken(test, allowedAuth, "/api/auth/config", "")

	if res.StatusCode != http.StatusOK {
		test.Fatalf("status = %d, want 200", res.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		test.Fatalf("decode: %v", err)
	}
	if body["neonAuthUrl"] != "https://auth.example.com" {
		test.Fatalf("neonAuthUrl = %q", body["neonAuthUrl"])
	}
}

// triste
func TestTheGateKeepsOutWhoCannotProveOrIsNotAllowed(test *testing.T) {
	cases := []struct {
		name   string
		cfg    Auth
		token  string
		status int
	}{
		{"no token", allowedAuth, "", http.StatusUnauthorized},
		{"invalid token", allowedAuth, "forged", http.StatusUnauthorized},
		{"valid but not allowed", allowedAuth, "stranger", http.StatusForbidden},
		{"login not configured", Auth{}, "owner", http.StatusServiceUnavailable},
	}

	for _, testCase := range cases {
		res := getWithToken(test, testCase.cfg, "/api/auth/me", testCase.token)

		if res.StatusCode != testCase.status {
			test.Errorf("%s: status = %d, want %d", testCase.name, res.StatusCode, testCase.status)
		}

		var body errorBody
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil || body.Message == "" {
			test.Errorf("%s: want a JSON error with a message, got %v", testCase.name, err)
		}
	}
}
