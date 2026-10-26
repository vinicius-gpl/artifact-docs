package server

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

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/store"
)

// fakeArtifacts guarda o que as rotas pediram e devolve o que o teste mandou.
type fakeArtifacts struct {
	failure     error
	gotOwner    string
	gotLabel    string
	gotFilter   store.ArtifactFilter
	gotPatch    store.ArtifactPatch
	artifact    store.Artifact
	listResults []store.Artifact
}

func (fake *fakeArtifacts) ListArtifacts(_ context.Context, ownerID string, filter store.ArtifactFilter) ([]store.Artifact, error) {
	fake.gotOwner, fake.gotFilter = ownerID, filter

	return fake.listResults, fake.failure
}

func (fake *fakeArtifacts) OpenArtifact(_ context.Context, ownerID, label string) (store.ArtifactDetail, error) {
	fake.gotOwner, fake.gotLabel = ownerID, label

	return store.ArtifactDetail{Artifact: fake.artifact, Versions: []store.ArtifactVersion{}}, fake.failure
}

func (fake *fakeArtifacts) PatchArtifact(_ context.Context, ownerID, label string, patch store.ArtifactPatch) (store.Artifact, error) {
	fake.gotOwner, fake.gotLabel, fake.gotPatch = ownerID, label, patch

	return fake.artifact, fake.failure
}

var sampleArtifact = store.Artifact{
	ID: "a1", Label: "abertura-cnpj", Title: "Abertura do CNPJ", Kind: "plan", Format: "md",
	URL: "/a/token/", CurrentVersion: 1, Tags: []string{},
	CreatedAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
	UpdatedAt: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC),
}

// call faz a requisição como o dono logado (o token "owner" do allowedAuth).
func call(test *testing.T, deps Dependencies, method, target, body string) *http.Response {
	test.Helper()

	deps.Auth = allowedAuth
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer owner")
	recorder := httptest.NewRecorder()
	New(slog.New(slog.NewTextHandler(io.Discard, nil)), builtSPA, deps).ServeHTTP(recorder, request)

	return recorder.Result()
}

func decodeInto(test *testing.T, response *http.Response, target any) {
	test.Helper()

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		test.Fatalf("decode: %v", err)
	}
}

// feliz
func TestListArtifactsFiltersByKindAndSearchForTheOwner(test *testing.T) {
	fake := &fakeArtifacts{listResults: []store.Artifact{sampleArtifact}}
	response := call(test, Dependencies{Artifacts: fake}, http.MethodGet, "/api/artifacts?kind=plan&q=cnpj", "")

	if response.StatusCode != http.StatusOK {
		test.Fatalf("status = %d, want 200", response.StatusCode)
	}
	if fake.gotOwner != "u1" || fake.gotFilter.Kind != "plan" || fake.gotFilter.Search != "cnpj" {
		test.Fatalf("owner = %q, filter = %+v", fake.gotOwner, fake.gotFilter)
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
func TestListArtifactsIgnoresAnUnknownKindLikeTheMock(test *testing.T) {
	fake := &fakeArtifacts{listResults: []store.Artifact{}}
	response := call(test, Dependencies{Artifacts: fake}, http.MethodGet, "/api/artifacts?kind=nope", "")

	if response.StatusCode != http.StatusOK || fake.gotFilter.Kind != "" {
		test.Fatalf("status = %d, kind = %q; want 200 and no kind filter", response.StatusCode, fake.gotFilter.Kind)
	}

	var body map[string]any
	decodeInto(test, response, &body)
	if items, ok := body["items"].([]any); !ok || len(items) != 0 {
		test.Fatalf("items = %v, want an empty list (never null)", body["items"])
	}
}

// feliz
func TestOpenArtifactAnswersTheDetailWithVersions(test *testing.T) {
	fake := &fakeArtifacts{artifact: sampleArtifact}
	response := call(test, Dependencies{Artifacts: fake}, http.MethodGet, "/api/artifacts/abertura-cnpj", "")

	if response.StatusCode != http.StatusOK || fake.gotLabel != "abertura-cnpj" {
		test.Fatalf("status = %d, label = %q", response.StatusCode, fake.gotLabel)
	}

	var body map[string]any
	decodeInto(test, response, &body)
	if body["title"] != "Abertura do CNPJ" || body["versions"] == nil {
		test.Fatalf("body = %v, want the artifact flattened with versions", body)
	}
}

// feliz
func TestPatchArtifactTellsMissingStatusFromNullStatus(test *testing.T) {
	cases := map[string]struct {
		body      string
		statusSet bool
	}{
		"status absent": {`{"isPinned":true}`, false},
		"status null":   {`{"isPinned":true,"status":null}`, true},
	}

	for name, testCase := range cases {
		fake := &fakeArtifacts{artifact: sampleArtifact}
		response := call(test, Dependencies{Artifacts: fake}, http.MethodPatch, "/api/artifacts/abertura-cnpj", testCase.body)

		if response.StatusCode != http.StatusOK {
			test.Fatalf("%s: status = %d, want 200", name, response.StatusCode)
		}
		if fake.gotPatch.IsPinned == nil || !*fake.gotPatch.IsPinned {
			test.Fatalf("%s: isPinned not passed", name)
		}
		if fake.gotPatch.StatusSet != testCase.statusSet || fake.gotPatch.Status != nil {
			test.Fatalf("%s: patch = %+v, want StatusSet %v and Status nil", name, fake.gotPatch, testCase.statusSet)
		}
	}
}

// triste
func TestArtifactRoutesAnswerTheRightError(test *testing.T) {
	cases := []struct {
		name    string
		deps    Dependencies
		method  string
		target  string
		body    string
		status  int
		message string
	}{
		{"not found", Dependencies{Artifacts: &fakeArtifacts{failure: store.ErrNotFound}}, http.MethodGet, "/api/artifacts/sumiu", "", http.StatusNotFound, "Artefato não encontrado"},
		{"patch not found", Dependencies{Artifacts: &fakeArtifacts{failure: store.ErrNotFound}}, http.MethodPatch, "/api/artifacts/sumiu", `{}`, http.StatusNotFound, "Artefato não encontrado"},
		{"broken body", Dependencies{Artifacts: &fakeArtifacts{}}, http.MethodPatch, "/api/artifacts/x", `{"isPinned":`, http.StatusBadRequest, "Alteração inválida"},
		{"wrong type", Dependencies{Artifacts: &fakeArtifacts{}}, http.MethodPatch, "/api/artifacts/x", `{"isPinned":"sim"}`, http.StatusBadRequest, "Alteração inválida"},
		{"unknown status", Dependencies{Artifacts: &fakeArtifacts{}}, http.MethodPatch, "/api/artifacts/x", `{"status":"lixo"}`, http.StatusBadRequest, "Situação inválida"},
		{"database down", Dependencies{Artifacts: &fakeArtifacts{failure: errors.New("connection refused")}}, http.MethodGet, "/api/artifacts", "", http.StatusInternalServerError, "Tente de novo"},
		{"no database", Dependencies{}, http.MethodGet, "/api/artifacts", "", http.StatusServiceUnavailable, "banco de dados"},
	}

	for _, testCase := range cases {
		response := call(test, testCase.deps, testCase.method, testCase.target, testCase.body)

		if response.StatusCode != testCase.status {
			test.Errorf("%s: status = %d, want %d", testCase.name, response.StatusCode, testCase.status)
			continue
		}

		var body errorBody
		decodeInto(test, response, &body)
		if !strings.Contains(body.Message, testCase.message) {
			test.Errorf("%s: message = %q, want it to contain %q", testCase.name, body.Message, testCase.message)
		}
		if strings.Contains(body.Message, "connection refused") {
			test.Errorf("%s: the technical cause leaked to the screen", testCase.name)
		}
	}
}
