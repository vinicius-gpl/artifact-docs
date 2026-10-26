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

func get(test *testing.T, handler http.Handler, target string) *http.Response {
	test.Helper()

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

	return rec.Result()
}

func newServer(spa fstest.MapFS) http.Handler {
	return New(slog.New(slog.NewTextHandler(io.Discard, nil)), spa, Auth{})
}

func bodyOf(test *testing.T, res *http.Response) string {
	test.Helper()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		test.Fatalf("read body: %v", err)
	}

	return string(raw)
}

// feliz
func TestHealthAnswersOK(test *testing.T) {
	res := get(test, newServer(builtSPA), "/api/health")

	if res.StatusCode != http.StatusOK {
		test.Fatalf("status = %d, want 200", res.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		test.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		test.Fatalf("status field = %q, want ok", body["status"])
	}
}

// triste
func TestUnknownAPIPathIsAJSONNotFound(test *testing.T) {
	res := get(test, newServer(builtSPA), "/api/nope")

	if res.StatusCode != http.StatusNotFound {
		test.Fatalf("status = %d, want 404", res.StatusCode)
	}
	if !strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
		test.Fatalf("content-type = %q, want JSON", res.Header.Get("Content-Type"))
	}
}

// feliz
func TestServesExistingSPAFiles(test *testing.T) {
	res := get(test, newServer(builtSPA), "/_app/immutable/app.js")

	if res.StatusCode != http.StatusOK {
		test.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if got := bodyOf(test, res); got != "console.log('app')" {
		test.Fatalf("body = %q", got)
	}
}

// feliz
func TestUnknownScreenFallsBackToIndex(test *testing.T) {
	for _, target := range []string{"/", "/artifacts", "/artifacts/briefing"} {
		res := get(test, newServer(builtSPA), target)

		if res.StatusCode != http.StatusOK {
			test.Fatalf("%s: status = %d, want 200", target, res.StatusCode)
		}
		if got := bodyOf(test, res); got != "<html>spa</html>" {
			test.Fatalf("%s: body = %q, want the index", target, got)
		}
		if res.Header.Get("Cache-Control") != "no-cache" {
			test.Fatalf("%s: index must not be cached", target)
		}
	}
}

// triste
func TestExplainsWhenTheSPAWasNotBuilt(test *testing.T) {
	res := get(test, newServer(fstest.MapFS{".gitkeep": {}}), "/artifacts")

	if res.StatusCode != http.StatusServiceUnavailable {
		test.Fatalf("status = %d, want 503", res.StatusCode)
	}
	if !strings.Contains(bodyOf(test, res), "mise run build") {
		test.Fatal("body must tell how to build the SPA")
	}
}
