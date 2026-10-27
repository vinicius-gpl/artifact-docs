package artifacts

import (
	"bytes"
	_ "embed"
	"html/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

// O Markdown vira HTML UMA vez, no upload, dentro do template da marca, e é salvo como
// index.html (briefing, módulo 1). Servir é só entregar o arquivo pronto: nenhum Markdown é
// convertido a cada visita.

//go:embed templates/markdown.html
var markdownPage string

var markdownTemplate = template.Must(template.New("markdown").Parse(markdownPage))

// markdownRenderer usa o GFM (tabela, tachado, lista de tarefas, link solto) e dá id aos
// títulos, para o link para uma seção (#instalacao) funcionar.
//
// HTML cru dentro do Markdown é descartado (o padrão do goldmark). Quem quiser HTML publica
// HTML; e, mesmo que passasse, o artefato é servido com sandbox.
var markdownRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

// renderMarkdown devolve a página completa, com título e estilo da marca.
func renderMarkdown(source []byte, title string) ([]byte, error) {
	var body bytes.Buffer
	if err := markdownRenderer.Convert(source, &body); err != nil {
		return nil, err
	}

	var page bytes.Buffer
	err := markdownTemplate.Execute(&page, struct {
		Title string
		Body  template.HTML
	}{
		Title: title,
		// O goldmark já escapou o texto e descartou o HTML cru: é seguro marcar como HTML.
		Body: template.HTML(body.String()),
	})
	if err != nil {
		return nil, err
	}

	return page.Bytes(), nil
}
