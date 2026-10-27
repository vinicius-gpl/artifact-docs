package apps

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

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/httpjson"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/database"
)

const validID = "6f1c2a3b-4d5e-4f60-8a7b-9c0d1e2f3a4b"

type fakeStore struct {
	hub      Hub
	failure  error
	gotID    string
	gotInput Input
	link     Link
}

func (store *fakeStore) Hub(_ context.Context) (Hub, error) {
	return store.hub, store.failure
}

func (store *fakeStore) Create(_ context.Context, input Input) (Link, error) {
	store.gotInput = input

	return store.link, store.failure
}

func (store *fakeStore) Update(_ context.Context, id string, input Input) (Link, error) {
	store.gotID, store.gotInput = id, input

	return store.link, store.failure
}

func (store *fakeStore) Delete(_ context.Context, id string) error {
	store.gotID = id

	return store.failure
}

type fakeChecker struct {
	checked []Target
}

func (checker *fakeChecker) CheckSoon(target Target) {
	checker.checked = append(checker.checked, target)
}

func newTestHandlers(store Store, checker Checker) *Handlers {
	return NewHandlers(store, checker, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func create(handlers *Handlers) http.HandlerFunc { return handlers.Create }
func update(handlers *Handlers) http.HandlerFunc { return handlers.Update }
func remove(handlers *Handlers) http.HandlerFunc { return handlers.Delete }
func hub(handlers *Handlers) http.HandlerFunc    { return handlers.Hub }

func serve(handler http.HandlerFunc, method, id, body string) *http.Response {
	request := httptest.NewRequest(method, "/api/apps", strings.NewReader(body))
	request.SetPathValue("id", id)
	recorder := httptest.NewRecorder()
	handler(recorder, request)

	return recorder.Result()
}

// feliz
func TestHubAnswersTypesAndApps(test *testing.T) {
	store := &fakeStore{hub: Hub{
		Types: []Type{{ID: "t1", Name: "Trabalho", Position: 0}},
		Apps:  []Link{{ID: "p1", TypeID: "t1", Name: "Neon", URL: "https://neon.tech", Status: "unknown"}},
	}}
	response := serve(newTestHandlers(store, nil).Hub, http.MethodGet, "", "")

	var body struct {
		Types []map[string]any `json:"types"`
		Apps  []map[string]any `json:"apps"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		test.Fatalf("decode: %v", err)
	}
	if response.StatusCode != http.StatusOK || len(body.Apps) != 1 || body.Apps[0]["typeId"] != "t1" {
		test.Fatalf("status = %d, body = %+v", response.StatusCode, body)
	}
	if logo, ok := body.Apps[0]["logoUrl"]; !ok || logo != nil {
		test.Fatalf("logoUrl = %v, want null (the contract requires the key)", logo)
	}
}

// feliz
func TestCreateCleansTheInputAndAsksForAQuickCheck(test *testing.T) {
	store := &fakeStore{link: Link{ID: validID, Name: "Jellyfin", URL: "https://jellyfin.casa", Status: "unknown"}}
	checker := &fakeChecker{}
	body := `{"name":"  Jellyfin ","url":" https://jellyfin.casa ","typeName":" Self-hosted "}`

	response := serve(newTestHandlers(store, checker).Create, http.MethodPost, "", body)

	if response.StatusCode != http.StatusCreated {
		test.Fatalf("status = %d, want 201", response.StatusCode)
	}
	if store.gotInput != (Input{Name: "Jellyfin", URL: "https://jellyfin.casa", TypeName: "Self-hosted"}) {
		test.Fatalf("input = %+v, want it trimmed", store.gotInput)
	}
	if len(checker.checked) != 1 || checker.checked[0].URL != "https://jellyfin.casa" {
		test.Fatalf("checked = %+v, want the new app checked right away", checker.checked)
	}
}

// feliz
func TestUpdateAndDeleteReachTheStoreByID(test *testing.T) {
	store := &fakeStore{link: Link{ID: validID}}
	handlers := newTestHandlers(store, nil)
	body := `{"name":"Jellyfin","url":"https://jellyfin.casa","typeName":"Mídia"}`

	updated := serve(handlers.Update, http.MethodPatch, validID, body)
	deleted := serve(handlers.Delete, http.MethodDelete, validID, "")

	if updated.StatusCode != http.StatusOK || deleted.StatusCode != http.StatusNoContent || store.gotID != validID {
		test.Fatalf("update = %d, delete = %d, id = %q", updated.StatusCode, deleted.StatusCode, store.gotID)
	}
}

// triste
func TestAppRoutesAnswerTheRightError(test *testing.T) {
	valid := `{"name":"Jellyfin","url":"https://jellyfin.casa","typeName":"Mídia"}`
	cases := []struct {
		name    string
		store   *fakeStore
		route   func(*Handlers) http.HandlerFunc
		id      string
		body    string
		status  int
		message string
	}{
		{"broken body", &fakeStore{}, create, "", `{"name":`, http.StatusBadRequest, "Cadastro inválido"},
		{"no name", &fakeStore{}, create, "", `{"name":" ","url":"https://x.y","typeName":"T"}`, http.StatusBadRequest, "nome"},
		{"not a web address", &fakeStore{}, create, "", `{"name":"X","url":"ftp://x.y","typeName":"T"}`, http.StatusBadRequest, "http://"},
		{"no type", &fakeStore{}, create, "", `{"name":"X","url":"https://x.y","typeName":""}`, http.StatusBadRequest, "tipo"},
		{"bad id", &fakeStore{}, update, "1 OR 1=1", valid, http.StatusNotFound, "App não encontrado"},
		{"missing app", &fakeStore{failure: database.ErrNotFound}, remove, validID, "", http.StatusNotFound, "App não encontrado"},
		{"database down", &fakeStore{failure: errors.New("connection refused")}, hub, "", "", http.StatusInternalServerError, "Tente de novo"},
	}

	for _, testCase := range cases {
		response := serve(testCase.route(newTestHandlers(testCase.store, nil)), http.MethodPost, testCase.id, testCase.body)

		if response.StatusCode != testCase.status {
			test.Errorf("%s: status = %d, want %d", testCase.name, response.StatusCode, testCase.status)
			continue
		}
		var body httpjson.ErrorBody
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil || !strings.Contains(body.Message, testCase.message) {
			test.Errorf("%s: message = %q, want it to contain %q", testCase.name, body.Message, testCase.message)
		}
	}
}
