package artifacts

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func serveRequest(handlers *Handlers, access, filePath string) *http.Response {
	request := httptest.NewRequest(http.MethodGet, "/a/"+access+"/"+filePath, nil)
	request.SetPathValue("access", access)
	request.SetPathValue("path", filePath)
	recorder := httptest.NewRecorder()
	handlers.Serve(recorder, request)

	return recorder.Result()
}

func servingHandlers(served Served) (*Handlers, *fakeFiles) {
	files := newFakeFiles()
	files.objects["artifacts/a1/v2/index.html"] = []byte("<h1>no ar</h1>")
	files.objects["artifacts/a1/v2/css/site.css"] = []byte("h1{}")
	files.objects["artifacts/a1/v2/index.pdf"] = []byte("%PDF-1.7")

	return newTestHandlers(&fakeStore{served: served}, files), files
}

// feliz
func TestServeDeliversTheCurrentVersionInsideTheSandbox(test *testing.T) {
	handlers, _ := servingHandlers(Served{ArtifactID: "a1", Version: 2, Format: "html"})

	response := serveRequest(handlers, "token", "")
	body, _ := io.ReadAll(response.Body)

	// A página de entrada ganha o cartão do link no começo; o conteúdo vem intacto depois.
	if response.StatusCode != http.StatusOK || !strings.HasSuffix(string(body), "<h1>no ar</h1>") {
		test.Fatalf("status = %d, body = %q", response.StatusCode, body)
	}
	if response.Header.Get("Content-Security-Policy") != sandboxPolicy {
		test.Fatalf("csp = %q, want the sandbox", response.Header.Get("Content-Security-Policy"))
	}
	if response.Header.Get("Referrer-Policy") != "no-referrer" || !strings.HasPrefix(response.Header.Get("ETag"), `"etag-1-`) {
		test.Fatalf("headers = %v", response.Header)
	}

	asset := serveRequest(handlers, "token", "css/site.css")
	if asset.StatusCode != http.StatusOK || !strings.HasPrefix(asset.Header.Get("Content-Type"), "text/css") {
		test.Fatalf("asset: status = %d, type = %q", asset.StatusCode, asset.Header.Get("Content-Type"))
	}
}

// feliz
func TestServeOpensAPDFWithoutTheSandbox(test *testing.T) {
	handlers, _ := servingHandlers(Served{ArtifactID: "a1", Version: 2, Format: "pdf"})

	response := serveRequest(handlers, "token", "")

	if response.StatusCode != http.StatusOK || response.Header.Get("Content-Type") != "application/pdf" {
		test.Fatalf("status = %d, type = %q", response.StatusCode, response.Header.Get("Content-Type"))
	}
	if response.Header.Get("Content-Security-Policy") != "" {
		test.Fatal("the PDF viewer refuses to open inside a sandbox")
	}
}

// feliz
func TestServeOpensAPrivateArtifactWithASignedLink(test *testing.T) {
	handlers, _ := servingHandlers(Served{ArtifactID: "a1", Version: 2, Format: "html", IsPrivate: true})
	signed := strings.Trim(strings.TrimPrefix(handlers.linker.URL("token", true), "/a/"), "/")

	response := serveRequest(handlers, signed, "")

	if response.StatusCode != http.StatusOK || response.Header.Get("Cache-Control") != "private, no-cache" {
		test.Fatalf("status = %d, cache = %q", response.StatusCode, response.Header.Get("Cache-Control"))
	}
}

// triste
func TestServeKeepsOutWhatIsPrivateHiddenOrMissing(test *testing.T) {
	private, _ := servingHandlers(Served{ArtifactID: "a1", Version: 2, Format: "html", IsPrivate: true})
	if response := serveRequest(private, "token", ""); response.StatusCode != http.StatusForbidden {
		test.Errorf("private without signature: status = %d, want 403", response.StatusCode)
	}

	public, _ := servingHandlers(Served{ArtifactID: "a1", Version: 2, Format: "html"})
	for _, filePath := range []string{"../v1/index.html", ".env", "nao-existe.css"} {
		if response := serveRequest(public, "token", filePath); response.StatusCode != http.StatusNotFound {
			test.Errorf("%s: status = %d, want 404", filePath, response.StatusCode)
		}
	}
}
