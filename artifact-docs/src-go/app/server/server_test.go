package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/neonauth"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/modules/auth"
)

var builtSPA = fstest.MapFS{"index.html": {Data: []byte("<html>spa</html>")}}

// acceptOwner aceita só o token "owner" — a criptografia tem teste em lib/infra/neonauth.
type acceptOwner struct{}

func (acceptOwner) Verify(_ context.Context, token string) (neonauth.User, error) {
	if token != "owner" {
		return neonauth.User{}, neonauth.ErrInvalidToken
	}

	return neonauth.User{ID: "u1", Email: "dono@exemplo.com", Name: "Dono"}, nil
}

var ownerAuth = auth.Settings{Tokens: acceptOwner{}, AllowedEmails: []string{"dono@exemplo.com"}}

func request(test *testing.T, deps Dependencies, method, target, token string) *http.Response {
	test.Helper()

	incoming := httptest.NewRequest(method, target, nil)
	if token != "" {
		incoming.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	New(slog.New(slog.NewTextHandler(io.Discard, nil)), builtSPA, deps).ServeHTTP(recorder, incoming)

	return recorder.Result()
}

// feliz
func TestHealthAnswersOK(test *testing.T) {
	response := request(test, Dependencies{}, http.MethodGet, "/api/health", "")

	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		test.Fatalf("decode: %v", err)
	}
	if response.StatusCode != http.StatusOK || body["status"] != "ok" {
		test.Fatalf("status = %d, body = %v", response.StatusCode, body)
	}
}

// feliz
func TestScreensFallBackToTheSPA(test *testing.T) {
	response := request(test, Dependencies{}, http.MethodGet, "/artifacts/briefing", "")

	raw, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusOK || string(raw) != "<html>spa</html>" {
		test.Fatalf("status = %d, body = %q, want the index", response.StatusCode, raw)
	}
}

// triste
func TestUnknownAPIPathIsAJSONNotFound(test *testing.T) {
	response := request(test, Dependencies{}, http.MethodGet, "/api/nope", "")

	if response.StatusCode != http.StatusNotFound {
		test.Fatalf("status = %d, want 404", response.StatusCode)
	}
	if !strings.HasPrefix(response.Header.Get("Content-Type"), "application/json") {
		test.Fatalf("content-type = %q, want JSON", response.Header.Get("Content-Type"))
	}
}

// triste
func TestProtectedRoutesPassThroughTheGate(test *testing.T) {
	for _, target := range []string{"/api/auth/me", "/api/artifacts", "/api/artifacts/x", "/api/apps"} {
		response := request(test, Dependencies{Auth: ownerAuth}, http.MethodGet, target, "")

		if response.StatusCode != http.StatusUnauthorized {
			test.Errorf("%s: status = %d, want 401 without a token", target, response.StatusCode)
		}
	}
}

// triste
func TestDataRoutesAnswer503WithoutADatabase(test *testing.T) {
	cases := []struct {
		method string
		target string
		token  string
	}{
		{http.MethodGet, "/api/auth/setup", ""},
		{http.MethodGet, "/api/artifacts", "owner"},
		{http.MethodGet, "/api/artifacts/x", "owner"},
		{http.MethodPatch, "/api/artifacts/x", "owner"},
		{http.MethodGet, "/api/apps", "owner"},
	}

	for _, testCase := range cases {
		response := request(test, Dependencies{Auth: ownerAuth}, testCase.method, testCase.target, testCase.token)

		if response.StatusCode != http.StatusServiceUnavailable {
			test.Errorf("%s %s: status = %d, want 503", testCase.method, testCase.target, response.StatusCode)
		}
	}
}
