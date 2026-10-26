package apps

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeStore struct {
	hub     Hub
	failure error
}

func (store fakeStore) Hub(_ context.Context) (Hub, error) {
	return store.hub, store.failure
}

func serveHub(store Store) *http.Response {
	recorder := httptest.NewRecorder()
	NewHandlers(store, slog.New(slog.NewTextHandler(io.Discard, nil))).
		Hub(recorder, httptest.NewRequest(http.MethodGet, "/api/apps", nil))

	return recorder.Result()
}

// feliz
func TestHubAnswersTypesAndApps(test *testing.T) {
	hub := Hub{
		Types: []Type{{ID: "t1", Name: "Trabalho", Position: 0}},
		Apps:  []Link{{ID: "p1", TypeID: "t1", Name: "Neon", URL: "https://neon.tech", Status: "unknown"}},
	}
	response := serveHub(fakeStore{hub: hub})

	if response.StatusCode != http.StatusOK {
		test.Fatalf("status = %d, want 200", response.StatusCode)
	}

	var body struct {
		Types []map[string]any `json:"types"`
		Apps  []map[string]any `json:"apps"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		test.Fatalf("decode: %v", err)
	}
	if len(body.Types) != 1 || len(body.Apps) != 1 || body.Apps[0]["typeId"] != "t1" {
		test.Fatalf("body = %+v", body)
	}
	if logo, ok := body.Apps[0]["logoUrl"]; !ok || logo != nil {
		test.Fatalf("logoUrl = %v, want null (the contract requires the key)", logo)
	}
}

// triste
func TestHubExplainsWhenItCannotAnswer(test *testing.T) {
	response := serveHub(fakeStore{failure: errors.New("timeout")})

	if response.StatusCode != http.StatusInternalServerError {
		test.Fatalf("status = %d, want 500", response.StatusCode)
	}
}
