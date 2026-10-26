package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

var builtSPA = fstest.MapFS{
	"index.html":            {Data: []byte("<html>spa</html>")},
	"_app/immutable/app.js": {Data: []byte("console.log('app')")},
}

func get(t *testing.T, handler http.Handler, target string) *http.Response {
	t.Helper()

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

	return rec.Result()
}

func newServer(spa fstest.MapFS) http.Handler {
	return New(slog.New(slog.NewTextHandler(io.Discard, nil)), spa, Auth{})
}

func bodyOf(t *testing.T, res *http.Response) string {
	t.Helper()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return string(raw)
}

// feliz
func TestHealthAnswersOK(t *testing.T) {
	res := get(t, newServer(builtSPA), "/api/health")

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field = %q, want ok", body["status"])
	}
}

// triste
func TestUnknownAPIPathIsAJSONNotFound(t *testing.T) {
	res := get(t, newServer(builtSPA), "/api/nope")

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}
	if !strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
		t.Fatalf("content-type = %q, want JSON", res.Header.Get("Content-Type"))
	}
}

// feliz
func TestServesExistingSPAFiles(t *testing.T) {
	res := get(t, newServer(builtSPA), "/_app/immutable/app.js")

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if got := bodyOf(t, res); got != "console.log('app')" {
		t.Fatalf("body = %q", got)
	}
}

// feliz
func TestUnknownScreenFallsBackToIndex(t *testing.T) {
	for _, target := range []string{"/", "/artifacts", "/artifacts/briefing"} {
		res := get(t, newServer(builtSPA), target)

		if res.StatusCode != http.StatusOK {
			t.Fatalf("%s: status = %d, want 200", target, res.StatusCode)
		}
		if got := bodyOf(t, res); got != "<html>spa</html>" {
			t.Fatalf("%s: body = %q, want the index", target, got)
		}
		if res.Header.Get("Cache-Control") != "no-cache" {
			t.Fatalf("%s: index must not be cached", target)
		}
	}
}

// triste
func TestExplainsWhenTheSPAWasNotBuilt(t *testing.T) {
	res := get(t, newServer(fstest.MapFS{".gitkeep": {}}), "/artifacts")

	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", res.StatusCode)
	}
	if !strings.Contains(bodyOf(t, res), "mise run build") {
		t.Fatal("body must tell how to build the SPA")
	}
}
