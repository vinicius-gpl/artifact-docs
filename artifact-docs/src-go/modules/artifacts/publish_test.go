package artifacts

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type formPart struct {
	field    string
	fileName string
	content  string
}

// publishRequest monta o PUT multipart que a tela e as skills mandam.
func publishRequest(test *testing.T, label string, parts ...formPart) *http.Request {
	test.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, part := range parts {
		var err error
		if part.fileName != "" {
			fileWriter, createErr := writer.CreateFormFile(part.field, part.fileName)
			err = createErr
			if err == nil {
				_, err = fileWriter.Write([]byte(part.content))
			}
		} else {
			err = writer.WriteField(part.field, part.content)
		}
		if err != nil {
			test.Fatalf("build form: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		test.Fatalf("close form: %v", err)
	}

	// A label vai escapada no endereço: uma label inválida com espaço ("Guia Rápido") é justamente
	// o que o teste manda, e crua ela quebra a linha da requisição antes de chegar ao handler.
	request := httptest.NewRequest(http.MethodPut, "/api/artifacts/"+url.PathEscape(label), &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.SetPathValue("label", label)

	return request
}

func publish(handlers *Handlers) http.HandlerFunc { return handlers.Publish }

const pngThumb = "\x89PNG\r\n\x1a\nfake-image"

// feliz
func TestPublishANewHTMLArtifactUploadsBeforeSaving(test *testing.T) {
	store := &fakeStore{artifact: sampleArtifact}
	files := newFakeFiles()
	request := publishRequest(test, "guia-rapido",
		formPart{field: "file", fileName: "guia.html", content: "<h1>Guia</h1><script>x()</script><p>contador</p>"},
		formPart{field: "thumb", fileName: "thumb.png", content: pngThumb},
		formPart{field: "title", content: "Guia rápido"},
		formPart{field: "kind", content: "document"},
		formPart{field: "tags", content: "hub, marca, hub"},
	)

	response := call(newTestHandlers(store, files), publish, request)

	if response.StatusCode != http.StatusCreated {
		test.Fatalf("status = %d, want 201", response.StatusCode)
	}
	record := store.gotRecord
	if !record.IsNew || record.Version != 1 || record.Format != "html" || !record.HasThumb {
		test.Fatalf("record = %+v", record)
	}
	if record.ContentText != "Guia contador" {
		test.Fatalf("content text = %q, want the visible text without the script", record.ContentText)
	}
	if strings.Join(record.Meta.Tags, ",") != "hub,marca" {
		test.Fatalf("tags = %v, want them without repetition", record.Meta.Tags)
	}

	folder := versionFolder(record.ArtifactID, 1)
	if _, ok := files.objects[folder+"index.html"]; !ok {
		test.Fatalf("index.html not uploaded; objects = %v", keys(files.objects))
	}
	if _, ok := files.objects[folder+thumbFile]; !ok {
		test.Fatal("thumb not uploaded")
	}
}

// feliz
func TestPublishANewVersionOfAnExistingArtifact(test *testing.T) {
	store := &fakeStore{artifact: sampleArtifact, target: PublishTarget{Exists: true, ArtifactID: "a1", LinkToken: "token", LatestVersion: 3}}
	files := newFakeFiles()
	request := publishRequest(test, "abertura-cnpj", formPart{field: "file", fileName: "notas.md", content: "# Abertura\n\nJunta comercial."})

	response := call(newTestHandlers(store, files), publish, request)

	if response.StatusCode != http.StatusOK {
		test.Fatalf("status = %d, want 200", response.StatusCode)
	}
	if store.gotRecord.IsNew || store.gotRecord.Version != 4 || store.gotRecord.Format != "md" {
		test.Fatalf("record = %+v, want version 4 of the existing artifact", store.gotRecord)
	}
	page := string(files.objects["artifacts/a1/v4/index.html"])
	if !strings.Contains(page, `<h1 id="abertura">Abertura</h1>`) || !strings.Contains(page, "IBM Plex Sans") {
		test.Fatalf("markdown page = %q, want the rendered body inside the brand template", page)
	}
}

// feliz
func TestPublishANewDocumentIsPrivateAndHasNoStatus(test *testing.T) {
	store := &fakeStore{artifact: sampleArtifact}
	request := publishRequest(test, "guia",
		formPart{field: "file", fileName: "guia.html", content: "<p>oi</p>"},
		formPart{field: "title", content: "Guia"},
		formPart{field: "kind", content: "document"},
		formPart{field: "status", content: "active"},
	)

	response := call(newTestHandlers(store, newFakeFiles()), publish, request)

	meta := store.gotRecord.Meta
	if response.StatusCode != http.StatusCreated || meta.IsPrivate == nil || !*meta.IsPrivate || meta.Status != nil {
		test.Fatalf("status = %d, meta = %+v, want a private document without status", response.StatusCode, meta)
	}
}

// feliz
func TestPublishANewIdeaTakesTheChosenStatusAndPrivacy(test *testing.T) {
	cases := []struct {
		name        string
		extra       []formPart
		wantStatus  string
		wantPrivate bool
	}{
		{"defaults", nil, "draft", true},
		{"chosen", []formPart{{field: "status", content: "active"}, {field: "private", content: "false"}}, "active", false},
	}

	for _, testCase := range cases {
		store := &fakeStore{artifact: sampleArtifact}
		parts := append([]formPart{
			{field: "file", fileName: "ideia.md", content: "# Ideia"},
			{field: "title", content: "Ideia"},
			{field: "kind", content: "idea"},
		}, testCase.extra...)

		response := call(newTestHandlers(store, newFakeFiles()), publish, publishRequest(test, "ideia", parts...))

		if response.StatusCode != http.StatusCreated {
			test.Fatalf("%s: status = %d, want 201", testCase.name, response.StatusCode)
		}
		meta := store.gotRecord.Meta
		if meta.Status == nil || *meta.Status != testCase.wantStatus || *meta.IsPrivate != testCase.wantPrivate {
			test.Errorf("%s: meta = %+v, want status %q and private %v", testCase.name, meta, testCase.wantStatus, testCase.wantPrivate)
		}
	}
}

// triste
func TestPublishRefusesWhatCannotBePublished(test *testing.T) {
	html := formPart{field: "file", fileName: "site.html", content: "<p>oi</p>"}
	cases := []struct {
		name    string
		label   string
		parts   []formPart
		message string
	}{
		{"bad label", "Guia Rápido", []formPart{html}, "letras minúsculas"},
		{"new without title", "guia", []formPart{html, {field: "kind", content: "document"}}, "título e tipo"},
		{"unknown kind", "guia", []formPart{html, {field: "kind", content: "video"}}, "document, plan ou idea"},
		{"unknown status", "guia", []formPart{html, {field: "status", content: "pronto"}}, "draft, active, done ou archived"},
		{"private not a boolean", "guia", []formPart{html, {field: "private", content: "sim"}}, "true ou false"},
		// Com título e tipo: artefato novo sem eles é recusado antes de o arquivo ser aberto.
		{"unknown extension", "guia", []formPart{{field: "file", fileName: "planilha.xlsx", content: "x"}, {field: "title", content: "Planilha"}, {field: "kind", content: "document"}}, ".html, .zip, .md ou .pdf"},
		{"fake pdf", "guia", []formPart{{field: "file", fileName: "doc.pdf", content: "MZ"}, {field: "title", content: "Doc"}, {field: "kind", content: "document"}}, "PDF válido"},
		{"thumb not png", "guia", []formPart{html, {field: "thumb", fileName: "thumb.jpg", content: "\xff\xd8"}}, "PNG"},
		{"no file", "guia", []formPart{{field: "title", content: "Guia"}}, "campo file"},
	}

	for _, testCase := range cases {
		files := newFakeFiles()
		response := call(newTestHandlers(&fakeStore{}, files), publish, publishRequest(test, testCase.label, testCase.parts...))

		if response.StatusCode != http.StatusBadRequest {
			test.Errorf("%s: status = %d, want 400", testCase.name, response.StatusCode)
			continue
		}
		var body struct{ Message string }
		decodeInto(test, response, &body)
		if !strings.Contains(body.Message, testCase.message) {
			test.Errorf("%s: message = %q, want it to contain %q", testCase.name, body.Message, testCase.message)
		}
		if len(files.objects) != 0 {
			test.Errorf("%s: nothing may reach the storage", testCase.name)
		}
	}
}

// triste
func TestPublishDiscardsTheUploadedFilesWhenSavingFails(test *testing.T) {
	store := &fakeStore{target: PublishTarget{Exists: true, ArtifactID: "a1", LatestVersion: 1}, saveFailure: ErrConflict}
	files := newFakeFiles()
	request := publishRequest(test, "abertura-cnpj", formPart{field: "file", fileName: "site.html", content: "<p>oi</p>"})

	response := call(newTestHandlers(store, files), publish, request)

	if response.StatusCode != http.StatusConflict {
		test.Fatalf("status = %d, want 409", response.StatusCode)
	}
	if len(files.deleted) != 1 || files.deleted[0] != "artifacts/a1/v2/" {
		test.Fatalf("deleted = %v, want the half-published version folder", files.deleted)
	}
}

// feliz
func TestNewUUIDIsAVersion4UUID(test *testing.T) {
	first, second := newUUID(), newUUID()

	if len(first) != 36 || first[14] != '4' || first == second {
		test.Fatalf("uuids = %q, %q", first, second)
	}
	if !strings.ContainsAny(first[19:20], "89ab") {
		test.Fatalf("variant = %q, want RFC 4122", first[19:20])
	}
}

func keys(objects map[string][]byte) []string {
	names := make([]string, 0, len(objects))
	for name := range objects {
		names = append(names, name)
	}

	return names
}
