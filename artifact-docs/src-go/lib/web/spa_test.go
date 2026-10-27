package web

import (
	"io"
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

func get(test *testing.T, dist fstest.MapFS, target string) (*http.Response, string) {
	test.Helper()

	recorder := httptest.NewRecorder()
	Handler(dist).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))

	response := recorder.Result()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		test.Fatalf("read body: %v", err)
	}

	return response, string(raw)
}

// feliz
func TestServesExistingSPAFiles(test *testing.T) {
	response, body := get(test, builtSPA, "/_app/immutable/app.js")

	if response.StatusCode != http.StatusOK {
		test.Fatalf("status = %d, want 200", response.StatusCode)
	}
	if body != "console.log('app')" {
		test.Fatalf("body = %q", body)
	}
}

// feliz
func TestUnknownScreenFallsBackToIndex(test *testing.T) {
	for _, target := range []string{"/", "/artifacts", "/artifacts/briefing"} {
		response, body := get(test, builtSPA, target)

		if response.StatusCode != http.StatusOK {
			test.Fatalf("%s: status = %d, want 200", target, response.StatusCode)
		}
		if body != "<html>spa</html>" {
			test.Fatalf("%s: body = %q, want the index", target, body)
		}
		if response.Header.Get("Cache-Control") != "no-cache" {
			test.Fatalf("%s: index must not be cached", target)
		}
	}
}

// triste
func TestExplainsWhenTheSPAWasNotBuilt(test *testing.T) {
	response, body := get(test, fstest.MapFS{".gitkeep": {}}, "/artifacts")

	if response.StatusCode != http.StatusServiceUnavailable {
		test.Fatalf("status = %d, want 503", response.StatusCode)
	}
	if !strings.Contains(body, "mise run build") {
		test.Fatal("body must tell how to build the SPA")
	}
}
