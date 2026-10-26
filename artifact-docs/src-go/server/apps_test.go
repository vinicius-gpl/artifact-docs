package server

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/store"
)

type fakeApps struct {
	hub     store.AppHub
	failure error
}

func (fake fakeApps) AppHub(_ context.Context) (store.AppHub, error) {
	return fake.hub, fake.failure
}

type fakeSetup struct {
	hasOwner  bool
	failure   error
	gotEmails []string
}

func (fake *fakeSetup) HasOwner(_ context.Context, allowedEmails []string) (bool, error) {
	fake.gotEmails = allowedEmails

	return fake.hasOwner, fake.failure
}

// feliz
func TestAppHubAnswersTypesAndApps(test *testing.T) {
	hub := store.AppHub{
		Types: []store.AppType{{ID: "t1", Name: "Trabalho", Position: 0}},
		Apps:  []store.AppLink{{ID: "p1", TypeID: "t1", Name: "Neon", URL: "https://neon.tech", Status: "unknown"}},
	}
	response := call(test, Dependencies{Apps: fakeApps{hub: hub}}, http.MethodGet, "/api/apps", "")

	if response.StatusCode != http.StatusOK {
		test.Fatalf("status = %d, want 200", response.StatusCode)
	}

	var body struct {
		Types []map[string]any `json:"types"`
		Apps  []map[string]any `json:"apps"`
	}
	decodeInto(test, response, &body)
	if len(body.Types) != 1 || len(body.Apps) != 1 || body.Apps[0]["typeId"] != "t1" {
		test.Fatalf("body = %+v", body)
	}
	if logo, ok := body.Apps[0]["logoUrl"]; !ok || logo != nil {
		test.Fatalf("logoUrl = %v, want null (the contract requires the key)", logo)
	}
}

// triste
func TestAppHubExplainsWhenItCannotAnswer(test *testing.T) {
	failing := call(test, Dependencies{Apps: fakeApps{failure: errors.New("timeout")}}, http.MethodGet, "/api/apps", "")
	if failing.StatusCode != http.StatusInternalServerError {
		test.Fatalf("store failure: status = %d, want 500", failing.StatusCode)
	}

	missing := call(test, Dependencies{}, http.MethodGet, "/api/apps", "")
	if missing.StatusCode != http.StatusServiceUnavailable {
		test.Fatalf("no database: status = %d, want 503", missing.StatusCode)
	}
}

// feliz
func TestSetupIsPublicAndOffersFirstAccessUntilTheOwnerExists(test *testing.T) {
	for _, hasOwner := range []bool{false, true} {
		setup := &fakeSetup{hasOwner: hasOwner}
		response := getWithDeps(test, Dependencies{Auth: allowedAuth, Setup: setup}, "/api/auth/setup")

		if response.StatusCode != http.StatusOK {
			test.Fatalf("status = %d, want 200", response.StatusCode)
		}

		var body map[string]bool
		decodeInto(test, response, &body)
		if body["needsSetup"] != !hasOwner {
			test.Fatalf("hasOwner %v: needsSetup = %v", hasOwner, body["needsSetup"])
		}
		if len(setup.gotEmails) != 1 || setup.gotEmails[0] != "dono@exemplo.com" {
			test.Fatalf("emails = %v, want the allowed list", setup.gotEmails)
		}
	}
}

// triste
func TestSetupExplainsWhenItCannotCheck(test *testing.T) {
	failing := getWithDeps(test, Dependencies{Setup: &fakeSetup{failure: errors.New("no neon_auth schema")}}, "/api/auth/setup")
	if failing.StatusCode != http.StatusInternalServerError {
		test.Fatalf("store failure: status = %d, want 500", failing.StatusCode)
	}

	missing := getWithDeps(test, Dependencies{}, "/api/auth/setup")
	if missing.StatusCode != http.StatusServiceUnavailable {
		test.Fatalf("no database: status = %d, want 503", missing.StatusCode)
	}
}
