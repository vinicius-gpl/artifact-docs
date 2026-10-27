package artifacts

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// zipUpload grava um .zip de verdade num temporário, com os arquivos dados.
func zipUpload(test *testing.T, entries map[string]string) upload {
	test.Helper()

	diskPath := filepath.Join(test.TempDir(), "site.zip")
	file, err := os.Create(diskPath)
	if err != nil {
		test.Fatalf("create: %v", err)
	}
	writer := zip.NewWriter(file)
	for name, content := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			test.Fatalf("zip entry: %v", err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			test.Fatalf("zip write: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		test.Fatalf("zip close: %v", err)
	}
	info, _ := file.Stat()
	_ = file.Close()

	return upload{diskPath: diskPath, name: "site.zip", size: info.Size()}
}

func paths(files []bundleFile) string {
	names := make([]string, 0, len(files))
	for _, file := range files {
		names = append(names, file.path)
	}

	return strings.Join(names, " ")
}

// feliz
func TestAZippedSiteInsideAFolderIsServedFromTheFolder(test *testing.T) {
	received := zipUpload(test, map[string]string{
		"site/index.html":      "<title>Guia</title><p>Olá &amp; bem-vindo</p>",
		"site/css/style.css":   "p{}",
		"site/.DS_Store":       "lixo",
		"__MACOSX/site/._junk": "lixo",
	})

	content, err := buildBundle(received, "Guia")
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = content.release() }()

	if content.format != "html" || !strings.Contains(paths(content.files), "index.html") || !strings.Contains(paths(content.files), "css/style.css") {
		test.Fatalf("files = %q, want the site without the folder prefix", paths(content.files))
	}
	if strings.Contains(paths(content.files), "DS_Store") || strings.Contains(paths(content.files), "MACOSX") {
		test.Fatalf("files = %q, hidden system files must be dropped", paths(content.files))
	}
	if content.text != "Guia Olá & bem-vindo" {
		test.Fatalf("text = %q", content.text)
	}
}

// triste
func TestAZipWithoutIndexOrWithAnEscapingPathIsRefused(test *testing.T) {
	cases := map[string]map[string]string{
		"no index":      {"pagina.html": "<p>x</p>"},
		"escaping path": {"index.html": "<p>x</p>", "../../etc/passwd": "x"},
		"two folders":   {"a/index.html": "<p>x</p>", "b/style.css": "p{}"},
	}

	for name, entries := range cases {
		_, err := buildBundle(zipUpload(test, entries), "Guia")

		var invalid invalidUpload
		if !errors.As(err, &invalid) {
			test.Errorf("%s: err = %v, want an invalidUpload", name, err)
		}
	}
}

// feliz
func TestHTMLTextKeepsWhatIsReadAndDropsScriptsAndStyles(test *testing.T) {
	page := `<html><head><style>p{color:red}</style><script type="module">alert(1)</script></head>
<body><h1>Abertura</h1><p>Contador &eacute; essencial</p><svg><text>ícone</text></svg></body></html>`

	if text := htmlText(page); text != "Abertura Contador é essencial" {
		test.Fatalf("text = %q", text)
	}
}

// triste
func TestHTMLTextSurvivesBrokenMarkup(test *testing.T) {
	for _, page := range []string{"<p>sem fim", "<script>nunca fecha", "texto < solto"} {
		_ = htmlText(page)
	}

	if text := htmlText("<p>sem fim"); text != "sem fim" {
		test.Fatalf("text = %q", text)
	}
}

// triste
func TestCleanTextCutsWithoutBreakingACharacter(test *testing.T) {
	text := cleanText(strings.Repeat("é", maxSearchText))

	if len(text) > maxSearchText || !strings.HasSuffix(text, "é") {
		test.Fatalf("len = %d, the cut must land on a character boundary", len(text))
	}
}
