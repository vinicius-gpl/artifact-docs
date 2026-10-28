package artifacts

import (
	"bytes"
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
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/storage"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/modules/auth"
)

// fakeStore guarda o que as rotas pediram e devolve o que o teste mandou.
type fakeStore struct {
	failure     error
	saveFailure error
	gotOwner    string
	gotLabel    string
	gotFilter   Filter
	gotPatch    Patch
	gotRecord   VersionRecord
	gotVersion  int
	artifact    Artifact
	found       []Artifact
	target      PublishTarget
	served      Served
}

func (store *fakeStore) List(_ context.Context, ownerID string, filter Filter) ([]Artifact, error) {
	store.gotOwner, store.gotFilter = ownerID, filter

	return store.found, store.failure
}

func (store *fakeStore) Open(_ context.Context, ownerID, label string) (Detail, error) {
	store.gotOwner, store.gotLabel = ownerID, label

	return Detail{Artifact: store.artifact, Versions: []Version{}}, store.failure
}

func (store *fakeStore) Find(_ context.Context, ownerID, label string) (Artifact, error) {
	store.gotOwner, store.gotLabel = ownerID, label

	return store.artifact, store.failure
}

func (store *fakeStore) Update(_ context.Context, ownerID, label string, patch Patch) (Artifact, error) {
	store.gotOwner, store.gotLabel, store.gotPatch = ownerID, label, patch

	return store.artifact, store.failure
}

func (store *fakeStore) FindServed(_ context.Context, _ string) (Served, error) {
	return store.served, store.failure
}

func (store *fakeStore) PublishTarget(_ context.Context, ownerID, label string) (PublishTarget, error) {
	store.gotOwner, store.gotLabel = ownerID, label

	return store.target, store.failure
}

func (store *fakeStore) SaveVersion(_ context.Context, record VersionRecord) error {
	store.gotRecord = record

	return store.saveFailure
}

func (store *fakeStore) Restore(_ context.Context, _, _ string, version int) (string, error) {
	store.gotVersion = version

	return "token", store.failure
}

func (store *fakeStore) Delete(_ context.Context, _, _ string) (Deleted, error) {
	return Deleted{ArtifactID: "a1", LinkToken: "token"}, store.failure
}

// fakeFiles é o R2 em memória.
type fakeFiles struct {
	objects map[string][]byte
	deleted []string
}

func newFakeFiles() *fakeFiles {
	return &fakeFiles{objects: map[string][]byte{}}
}

func (files *fakeFiles) Put(_ context.Context, key string, content io.Reader, size int64, _ string) error {
	raw, err := io.ReadAll(content)
	if err != nil {
		return err
	}
	if int64(len(raw)) != size {
		return errors.New("size mismatch")
	}
	files.objects[key] = raw

	return nil
}

func (files *fakeFiles) Get(_ context.Context, key string) (storage.Object, error) {
	raw, ok := files.objects[key]
	if !ok {
		return storage.Object{}, storage.ErrNotFound
	}

	return storage.Object{
		Content:      nopSeekCloser{bytes.NewReader(raw)},
		Size:         int64(len(raw)),
		ETag:         "etag-1",
		LastModified: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
	}, nil
}

func (files *fakeFiles) DeleteFolder(_ context.Context, prefix string) error {
	files.deleted = append(files.deleted, prefix)

	return nil
}

type nopSeekCloser struct {
	*bytes.Reader
}

func (nopSeekCloser) Close() error { return nil }

var sampleArtifact = Artifact{
	ID: "a1", Label: "abertura-cnpj", Title: "Abertura do CNPJ", Kind: "plan", Format: "md",
	CurrentVersion: 1, Tags: []string{}, linkToken: "token",
	CreatedAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
	UpdatedAt: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC),
}

func newTestHandlers(store Store, files Files) *Handlers {
	return NewHandlers(store, files, NewLinker([]byte("test-secret")), slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// call chama a rota pelo porteiro em modo de teste: entra o auth.TestUser, sem token.
func call(handlers *Handlers, route func(*Handlers) http.HandlerFunc, request *http.Request) *http.Response {
	recorder := httptest.NewRecorder()
	auth.RequireUser(auth.Settings{TestMode: true}, route(handlers))(recorder, request)

	return recorder.Result()
}

func jsonRequest(method, target, label, body string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.SetPathValue("label", label)

	return request
}

func list(handlers *Handlers) http.HandlerFunc    { return handlers.List }
func open(handlers *Handlers) http.HandlerFunc    { return handlers.Open }
func update(handlers *Handlers) http.HandlerFunc  { return handlers.Update }
func restore(handlers *Handlers) http.HandlerFunc { return handlers.Restore }
func remove(handlers *Handlers) http.HandlerFunc  { return handlers.Delete }

func decodeInto(test *testing.T, response *http.Response, target any) {
	test.Helper()

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		test.Fatalf("decode: %v", err)
	}
}

// feliz
func TestListFiltersByKindAndSearchForTheLoggedUser(test *testing.T) {
	store := &fakeStore{found: []Artifact{sampleArtifact}}
	response := call(newTestHandlers(store, nil), list, jsonRequest(http.MethodGet, "/api/artifacts?kind=plan&q=cnpj", "", ""))

	if response.StatusCode != http.StatusOK {
		test.Fatalf("status = %d, want 200", response.StatusCode)
	}
	if store.gotOwner != auth.TestUser.ID || store.gotFilter.Kind != "plan" || store.gotFilter.Search != "cnpj" {
		test.Fatalf("owner = %q, filter = %+v", store.gotOwner, store.gotFilter)
	}

	var body struct {
		Items []map[string]any `json:"items"`
	}
	decodeInto(test, response, &body)
	if len(body.Items) != 1 || body.Items[0]["url"] != "/a/token/" {
		test.Fatalf("items = %v, want the served link built by the Linker", body.Items)
	}
	if body.Items[0]["updatedAt"] != "2026-10-02T00:00:00Z" {
		test.Fatalf("updatedAt = %v, want UTC ending in Z", body.Items[0]["updatedAt"])
	}
}

// feliz
func TestListIgnoresAnUnknownKindLikeTheMock(test *testing.T) {
	store := &fakeStore{found: []Artifact{}}
	response := call(newTestHandlers(store, nil), list, jsonRequest(http.MethodGet, "/api/artifacts?kind=nope", "", ""))

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
	response := call(newTestHandlers(store, nil), open, jsonRequest(http.MethodGet, "/api/artifacts/abertura-cnpj", "abertura-cnpj", ""))

	var body map[string]any
	decodeInto(test, response, &body)
	if response.StatusCode != http.StatusOK || body["title"] != "Abertura do CNPJ" || body["versions"] == nil {
		test.Fatalf("status = %d, body = %v", response.StatusCode, body)
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
		response := call(newTestHandlers(store, nil), update, jsonRequest(http.MethodPatch, "/api/artifacts/x", "x", testCase.body))

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

// feliz
func TestUpdateEditsTitleDescriptionAndTagsWithThePublishRules(test *testing.T) {
	store := &fakeStore{artifact: sampleArtifact}
	body := `{"title":"  ERP pessoal  ","description":"","tags":[" erp ","financeiro","erp",""]}`
	response := call(newTestHandlers(store, nil), update, jsonRequest(http.MethodPatch, "/api/artifacts/x", "x", body))

	if response.StatusCode != http.StatusOK {
		test.Fatalf("status = %d, want 200", response.StatusCode)
	}
	patch := store.gotPatch
	if patch.Title == nil || *patch.Title != "ERP pessoal" {
		test.Fatalf("title = %v, want trimmed", patch.Title)
	}
	if !patch.DescriptionSet || patch.Description == nil || *patch.Description != "" {
		test.Fatalf("description = %+v, want set and empty (clears)", patch)
	}
	if strings.Join(patch.Tags, ",") != "erp,financeiro" {
		test.Fatalf("tags = %v, want trimmed, without blanks nor repeats", patch.Tags)
	}
}

// feliz
func TestUpdateKeepsWhatDidNotCome(test *testing.T) {
	store := &fakeStore{artifact: sampleArtifact}
	call(newTestHandlers(store, nil), update, jsonRequest(http.MethodPatch, "/api/artifacts/x", "x", `{"isPrivate":true}`))

	patch := store.gotPatch
	if patch.Title != nil || patch.DescriptionSet || patch.Tags != nil {
		test.Fatalf("patch = %+v, want title, description and tags untouched", patch)
	}
}

// triste
func TestUpdateRefusesWhatPublishWouldRefuse(test *testing.T) {
	cases := map[string]string{
		"empty title":    `{"title":"   "}`,
		"too many tags":  `{"tags":["a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u"]}`,
		"unknown status": `{"status":"lost"}`,
	}

	for name, body := range cases {
		store := &fakeStore{artifact: sampleArtifact}
		response := call(newTestHandlers(store, nil), update, jsonRequest(http.MethodPatch, "/api/artifacts/x", "x", body))

		if response.StatusCode != http.StatusBadRequest {
			test.Fatalf("%s: status = %d, want 400", name, response.StatusCode)
		}
		if store.gotLabel != "" {
			test.Fatalf("%s: the store was called with an invalid patch", name)
		}
	}
}

// feliz
func TestRestorePointsBackToAnOlderVersion(test *testing.T) {
	store := &fakeStore{artifact: sampleArtifact}
	response := call(newTestHandlers(store, nil), restore, jsonRequest(http.MethodPost, "/api/artifacts/x/restore", "x", `{"version":2}`))

	if response.StatusCode != http.StatusOK || store.gotVersion != 2 {
		test.Fatalf("status = %d, version = %d; want 200 and 2", response.StatusCode, store.gotVersion)
	}
}

// feliz
func TestDeleteRemovesTheRecordAndTheFiles(test *testing.T) {
	files := newFakeFiles()
	response := call(newTestHandlers(&fakeStore{}, files), remove, jsonRequest(http.MethodDelete, "/api/artifacts/x", "x", ""))

	if response.StatusCode != http.StatusNoContent {
		test.Fatalf("status = %d, want 204", response.StatusCode)
	}
	if len(files.deleted) != 1 || files.deleted[0] != "artifacts/a1/" {
		test.Fatalf("deleted = %v, want the artifact folder", files.deleted)
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
		{"restore without version", &fakeStore{}, restore, `{}`, http.StatusBadRequest, "Informe a versão"},
		{"restore missing version", &fakeStore{failure: database.ErrNotFound}, restore, `{"version":9}`, http.StatusNotFound, "Versão não encontrada"},
		{"database down", &fakeStore{failure: errors.New("connection refused")}, list, "", http.StatusInternalServerError, "Tente de novo"},
	}

	for _, testCase := range cases {
		response := call(newTestHandlers(testCase.store, nil), testCase.route, jsonRequest(http.MethodPost, "/api/artifacts/x", "x", testCase.body))

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
