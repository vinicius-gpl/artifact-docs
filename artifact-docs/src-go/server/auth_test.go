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

func (f fakeTokens) Verify(_ context.Context, token string) (auth.User, error) {
	user, ok := f[token]
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

func getWithToken(t *testing.T, cfg Auth, target, token string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, target, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	New(slog.New(slog.NewTextHandler(io.Discard, nil)), builtSPA, cfg).ServeHTTP(rec, req)

	return rec.Result()
}

// feliz
func TestMeAnswersWhoIsLoggedIn(t *testing.T) {
	res := getWithToken(t, allowedAuth, "/api/auth/me", "owner")

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != "u1" || body["email"] != "Dono@Exemplo.com" || body["name"] != "Dono" {
		t.Fatalf("body = %v", body)
	}
	if image, ok := body["image"]; !ok || image != nil {
		t.Fatalf("image = %v, want null (the contract requires the key)", image)
	}
}

// feliz
func TestAuthConfigIsPublic(t *testing.T) {
	res := getWithToken(t, allowedAuth, "/api/auth/config", "")

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["neonAuthUrl"] != "https://auth.example.com" {
		t.Fatalf("neonAuthUrl = %q", body["neonAuthUrl"])
	}
}

// triste
func TestTheGateKeepsOutWhoCannotProveOrIsNotAllowed(t *testing.T) {
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

	for _, tc := range cases {
		res := getWithToken(t, tc.cfg, "/api/auth/me", tc.token)

		if res.StatusCode != tc.status {
			t.Errorf("%s: status = %d, want %d", tc.name, res.StatusCode, tc.status)
		}

		var body errorBody
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil || body.Message == "" {
			t.Errorf("%s: want a JSON error with a message, got %v", tc.name, err)
		}
	}
}
