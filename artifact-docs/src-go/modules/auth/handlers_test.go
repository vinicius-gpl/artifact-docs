package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/neonauth"
)

type fakeOwners struct {
	hasOwner  bool
	failure   error
	gotEmails []string
}

func (owners *fakeOwners) HasOwner(_ context.Context, allowedEmails []string) (bool, error) {
	owners.gotEmails = allowedEmails

	return owners.hasOwner, owners.failure
}

func newTestHandlers(owners OwnerChecker) *Handlers {
	return NewHandlers(ownerSettings, owners, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func serve(handler http.HandlerFunc, request *http.Request) *http.Response {
	recorder := httptest.NewRecorder()
	handler(recorder, request)

	return recorder.Result()
}

// feliz
func TestMeAnswersWhoIsLoggedIn(test *testing.T) {
	request := WithUser(httptest.NewRequest(http.MethodGet, "/api/auth/me", nil),
		neonauth.User{ID: "u1", Email: "dono@exemplo.com", Name: "Dono"})
	response := serve(newTestHandlers(&fakeOwners{}).Me, request)

	var body map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		test.Fatalf("decode: %v", err)
	}
	if body["id"] != "u1" || body["email"] != "dono@exemplo.com" || body["name"] != "Dono" {
		test.Fatalf("body = %v", body)
	}
	if image, ok := body["image"]; !ok || image != nil {
		test.Fatalf("image = %v, want null (the contract requires the key)", image)
	}
}

// feliz
func TestConfigTellsWhereTheLoginHappens(test *testing.T) {
	response := serve(newTestHandlers(&fakeOwners{}).Config, httptest.NewRequest(http.MethodGet, "/api/auth/config", nil))

	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		test.Fatalf("decode: %v", err)
	}
	if body["neonAuthUrl"] != "https://auth.example.com" {
		test.Fatalf("neonAuthUrl = %q", body["neonAuthUrl"])
	}
}

// feliz
func TestSetupOffersFirstAccessUntilTheOwnerExists(test *testing.T) {
	for _, hasOwner := range []bool{false, true} {
		owners := &fakeOwners{hasOwner: hasOwner}
		response := serve(newTestHandlers(owners).Setup, httptest.NewRequest(http.MethodGet, "/api/auth/setup", nil))

		var body map[string]bool
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			test.Fatalf("decode: %v", err)
		}
		if body["needsSetup"] != !hasOwner {
			test.Fatalf("hasOwner %v: needsSetup = %v", hasOwner, body["needsSetup"])
		}
		if len(owners.gotEmails) != 1 || owners.gotEmails[0] != "dono@exemplo.com" {
			test.Fatalf("emails = %v, want the allowed list", owners.gotEmails)
		}
	}
}

// triste
func TestSetupExplainsWhenItCannotCheck(test *testing.T) {
	owners := &fakeOwners{failure: errors.New("relation neon_auth.user does not exist")}
	response := serve(newTestHandlers(owners).Setup, httptest.NewRequest(http.MethodGet, "/api/auth/setup", nil))

	if response.StatusCode != http.StatusInternalServerError {
		test.Fatalf("status = %d, want 500", response.StatusCode)
	}
}
