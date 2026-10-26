package artifacts

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/httpjson"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/database"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/neonauth"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/modules/auth"
)

// fakeStore guarda o que as rotas pediram e devolve o que o teste mandou.
type fakeStore struct {
	failure   error
	gotOwner  string
	gotLabel  string
	gotFilter Filter
	gotPatch  Patch
	artifact  Artifact
	found     []Artifact
}

func (store *fakeStore) List(_ context.Context, ownerID string, filter Filter) ([]Artifact, error) {
	store.gotOwner, store.gotFilter = ownerID, filter

	return store.found, store.failure
}

func (store *fakeStore) Open(_ context.Context, ownerID, label string) (Detail, error) {
	store.gotOwner, store.gotLabel = ownerID, label

	return Detail{Artifact: store.artifact, Versions: []Version{}}, store.failure
}

func (store *fakeStore) Update(_ context.Context, ownerID, label string, patch Patch) (Artifact, error) {
	store.gotOwner, store.gotLabel, store.gotPatch = ownerID, label, patch

	return store.artifact, store.failure
}

var sampleArtifact = Artifact{
	ID: "a1", Label: "abertura-cnpj", Title: "Abertura do CNPJ", Kind: "plan", Format: "md",
	URL: "/a/token/", CurrentVersion: 1, Tags: []string{},
	CreatedAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
	UpdatedAt: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC),
}

// call chama a rota como o dono logado — o porteiro tem teste próprio em modules/auth.
func call(store Store, route func(*Handlers) http.HandlerFunc, method, target, label, body string) *http.Response {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.SetPathValue("label", label)
	request = auth.WithUser(request, neonauth.User{ID: "u1", Email: "dono@exemplo.com", Name: "Dono"})

	recorder := httptest.NewRecorder()
	route(NewHandlers(store, slog.New(slog.NewTextHandler(io.Discard, nil))))(recorder, request)

	return recorder.Result()
}

func list(handlers *Handlers) http.HandlerFunc   { return handlers.List }
func open(handlers *Handlers) http.HandlerFunc   { return handlers.Open }
func update(handlers *Handlers) http.HandlerFunc { return handlers.Update }

func decodeInto(test *testing.T, response *http.Response, target any) {
	test.Helper()

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		test.Fatalf("decode: %v", err)
	}
}

// feliz
func TestListFiltersByKindAndSearchForTheOwner(test *testing.T) {
	store := &fakeStore{found: []Artifact{sampleArtifact}}
	response := call(store, list, http.MethodGet, "/api/artifacts?kind=plan&q=cnpj", "", "")

	if response.StatusCode != http.StatusOK {
		test.Fatalf("status = %d, want 200", response.StatusCode)
	}
	if store.gotOwner != "u1" || store.gotFilter.Kind != "plan" || store.gotFilter.Search != "cnpj" {
		test.Fatalf("owner = %q, filter = %+v", store.gotOwner, store.gotFilter)
	}

	var body struct {
		Items []map[string]any `json:"items"`
	}
	decodeInto(test, response, &body)
	if len(body.Items) != 1 || body.Items[0]["label"] != "abertura-cnpj" {
		test.Fatalf("items = %v", body.Items)
	}
	if body.Items[0]["updatedAt"] != "2026-10-02T00:00:00Z" {
		test.Fatalf("updatedAt = %v, want UTC ending in Z", body.Items[0]["updatedAt"])
	}
}

// feliz
func TestListIgnoresAnUnknownKindLikeTheMock(test *testing.T) {
	store := &fakeStore{found: []Artifact{}}
	response := call(store, list, http.MethodGet, "/api/artifacts?kind=nope", "", "")

	if response.StatusCode != http.StatusOK || store.gotFilter.Kind != "" {
		test.Fatalf("status = %d, kind = %q; want 200 and no kind filter", response.StatusCode, store.gotFilter.Kind)
	}

	var body map[string]any
	decodeInto(test, response, &body)
	if items, ok := body["items"].([]any); !ok || len(items) != 0 {
		test.Fatalf("items = %v, want an empty list (never null)", body["items"])
	}
}

// feliz
func TestOpenAnswersTheDetailWithVersions(test *testing.T) {
	store := &fakeStore{artifact: sampleArtifact}
	response := call(store, open, http.MethodGet, "/api/artifacts/abertura-cnpj", "abertura-cnpj", "")

	if response.StatusCode != http.StatusOK || store.gotLabel != "abertura-cnpj" {
		test.Fatalf("status = %d, label = %q", response.StatusCode, store.gotLabel)
	}

	var body map[string]any
	decodeInto(test, response, &body)
	if body["title"] != "Abertura do CNPJ" || body["versions"] == nil {
		test.Fatalf("body = %v, want the artifact flattened with versions", body)
	}
}

// feliz
func TestUpdateTellsMissingStatusFromNullStatus(test *testing.T) {
	cases := map[string]struct {
		body      string
		statusSet bool
	}{
		"status absent": {`{"isPinned":true}`, false},
		"status null":   {`{"isPinned":true,"status":null}`, true},
	}

	for name, testCase := range cases {
		store := &fakeStore{artifact: sampleArtifact}
		response := call(store, update, http.MethodPatch, "/api/artifacts/abertura-cnpj", "abertura-cnpj", testCase.body)

		if response.StatusCode != http.StatusOK {
			test.Fatalf("%s: status = %d, want 200", name, response.StatusCode)
		}
		if store.gotPatch.IsPinned == nil || !*store.gotPatch.IsPinned {
			test.Fatalf("%s: isPinned not passed", name)
		}
		if store.gotPatch.StatusSet != testCase.statusSet || store.gotPatch.Status != nil {
			test.Fatalf("%s: patch = %+v, want StatusSet %v and Status nil", name, store.gotPatch, testCase.statusSet)
		}
	}
}

// triste
func TestArtifactRoutesAnswerTheRightError(test *testing.T) {
	cases := []struct {
		name    string
		store   *fakeStore
		route   func(*Handlers) http.HandlerFunc
		body    string
		status  int
		message string
	}{
		{"open not found", &fakeStore{failure: database.ErrNotFound}, open, "", http.StatusNotFound, "Artefato não encontrado"},
		{"update not found", &fakeStore{failure: database.ErrNotFound}, update, `{}`, http.StatusNotFound, "Artefato não encontrado"},
		{"broken body", &fakeStore{}, update, `{"isPinned":`, http.StatusBadRequest, "Alteração inválida"},
		{"wrong type", &fakeStore{}, update, `{"isPinned":"sim"}`, http.StatusBadRequest, "Alteração inválida"},
		{"unknown status", &fakeStore{}, update, `{"status":"lixo"}`, http.StatusBadRequest, "Situação inválida"},
		{"database down", &fakeStore{failure: errors.New("connection refused")}, list, "", http.StatusInternalServerError, "Tente de novo"},
	}

	for _, testCase := range cases {
		response := call(testCase.store, testCase.route, http.MethodPatch, "/api/artifacts/x", "x", testCase.body)

		if response.StatusCode != testCase.status {
			test.Errorf("%s: status = %d, want %d", testCase.name, response.StatusCode, testCase.status)
			continue
		}

		var body httpjson.ErrorBody
		decodeInto(test, response, &body)
		if !strings.Contains(body.Message, testCase.message) {
			test.Errorf("%s: message = %q, want it to contain %q", testCase.name, body.Message, testCase.message)
		}
		if strings.Contains(body.Message, "connection refused") {
			test.Errorf("%s: the technical cause leaked to the screen", testCase.name)
		}
	}
}
