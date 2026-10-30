package artifacts

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func publicIdea() Served {
	description := `Juntar finanças "PF" & PJ`

	return Served{ArtifactID: "a1", Version: 2, Format: "html", Title: "ERP <pessoal>", Description: &description, Kind: "idea"}
}

func entryPage(test *testing.T, served Served, page string) string {
	test.Helper()
	handlers, files := servingHandlers(served)
	files.objects["artifacts/a1/v2/index.html"] = []byte(page)

	request := httptest.NewRequest(http.MethodGet, "/a/token/", nil)
	request.Host = "hub.exemplo.com"
	request.Header.Set("X-Forwarded-Proto", "https")
	request.SetPathValue("access", "token")
	request.SetPathValue("path", "")
	recorder := httptest.NewRecorder()
	handlers.Serve(recorder, request)

	body, _ := io.ReadAll(recorder.Result().Body)

	return string(body)
}

// feliz
func TestPublicPageGetsTheLinkCardRightAfterTheHead(test *testing.T) {
	body := entryPage(test, publicIdea(), `<!doctype html><html><head lang="pt"><title>Minha</title></head><body>oi</body></html>`)

	afterHead := body[strings.Index(body, `<head lang="pt">`)+len(`<head lang="pt">`):]
	if !strings.HasPrefix(afterHead, `<meta name="robots" content="noindex">`) {
		test.Fatalf("the card must come first in the head: %q", body)
	}
	for _, want := range []string{
		`<meta property="og:title" content="ERP &lt;pessoal&gt;">`,
		`<meta property="og:description" content="Juntar finanças &#34;PF&#34; &amp; PJ">`,
		`<meta property="og:url" content="https://hub.exemplo.com/a/token/">`,
		`<meta property="og:image" content="https://hub.exemplo.com/og-card.png">`,
		`<meta property="og:image:width" content="1200">`,
		`<meta name="twitter:card" content="summary_large_image">`,
	} {
		if !strings.Contains(body, want) {
			test.Errorf("missing %s in %q", want, body)
		}
	}
	if strings.Count(body, "<title>") != 1 {
		test.Errorf("the page already had a title, it must not get a second one: %q", body)
	}
}

// feliz
func TestPageWithoutHeadGetsOneAndTheThumbIsTheImage(test *testing.T) {
	served := publicIdea()
	served.HasThumb = true
	served.Description = nil

	body := entryPage(test, served, `<h1>sem head</h1>`)

	if !strings.HasPrefix(body, `<head><meta name="robots"`) || !strings.HasSuffix(body, `</head><h1>sem head</h1>`) {
		test.Fatalf("a page without head must get one at the start: %q", body)
	}
	for _, want := range []string{
		`<meta property="og:image" content="https://hub.exemplo.com/a/token/.thumb.png">`,
		`<meta property="og:description" content="Ideia no artifact-docs">`,
		`<title>ERP &lt;pessoal&gt;</title>`,
	} {
		if !strings.Contains(body, want) {
			test.Errorf("missing %s in %q", want, body)
		}
	}
}

// feliz
func TestMarkdownPageGetsTheCard(test *testing.T) {
	page, err := renderMarkdown([]byte("# Plano\n\ntexto"), "Plano")
	if err != nil {
		test.Fatal(err)
	}

	body := entryPage(test, Served{ArtifactID: "a1", Version: 2, Format: "md", Title: "Plano", Kind: "plan"}, string(page))

	if !strings.Contains(body, `<meta property="og:title" content="Plano">`) {
		test.Fatalf("markdown page without the card: %q", body)
	}
}

// triste
func TestPrivatePageGetsNoCardButStaysOutOfSearch(test *testing.T) {
	served := publicIdea()
	served.IsPrivate = true
	handlers, files := servingHandlers(served)
	files.objects["artifacts/a1/v2/index.html"] = []byte(`<html><head></head><body>segredo</body></html>`)
	signed := strings.Trim(strings.TrimPrefix(handlers.linker.URL("token", true), "/a/"), "/")

	response := serveRequest(handlers, signed, "")
	body, _ := io.ReadAll(response.Body)

	if !strings.Contains(string(body), `<meta name="robots" content="noindex">`) {
		test.Fatalf("missing noindex: %q", body)
	}
	if strings.Contains(string(body), "og:") || strings.Contains(string(body), "ERP") {
		test.Fatalf("a private page must not leak its title in a card: %q", body)
	}
}

// triste
func TestAssetsAndHeaderTagsAreNotMistakenForTheEntryPage(test *testing.T) {
	handlers, _ := servingHandlers(publicIdea())
	asset := serveRequest(handlers, "token", "css/site.css")
	body, _ := io.ReadAll(asset.Body)
	if string(body) != "h1{}" {
		test.Fatalf("assets go untouched, got %q", body)
	}

	page := injectPreview([]byte(`<header>topo</header>`), publicIdea(), previewLink{origin: "https://hub", linkToken: "token"})
	if !strings.HasPrefix(string(page), "<head><meta") {
		test.Fatalf("<header> is not <head>: %q", page)
	}
}
