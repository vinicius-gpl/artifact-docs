package httpjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// feliz
func TestWriteAnswersJSON(test *testing.T) {
	recorder := httptest.NewRecorder()
	Write(recorder, http.StatusCreated, map[string]string{"status": "ok"})

	if recorder.Code != http.StatusCreated {
		test.Fatalf("status = %d, want 201", recorder.Code)
	}
	if !strings.HasPrefix(recorder.Header().Get("Content-Type"), "application/json") {
		test.Fatalf("content-type = %q", recorder.Header().Get("Content-Type"))
	}
	if strings.TrimSpace(recorder.Body.String()) != `{"status":"ok"}` {
		test.Fatalf("body = %q", recorder.Body.String())
	}
}

// triste
func TestFailLogsTheCauseAndHidesItFromTheScreen(test *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/apps", nil)

	Fail(logger, recorder, request, errors.New("connection refused"), "Tente de novo.")

	if recorder.Code != http.StatusInternalServerError {
		test.Fatalf("status = %d, want 500", recorder.Code)
	}

	var body ErrorBody
	if err := json.NewDecoder(io.Reader(recorder.Body)).Decode(&body); err != nil {
		test.Fatalf("decode: %v", err)
	}
	if body.Message != "Tente de novo." {
		test.Fatalf("message = %q", body.Message)
	}
	if !strings.Contains(logs.String(), "connection refused") {
		test.Fatal("the technical cause must reach the log")
	}
}
