package artifacts

import (
	"html"
	"strings"
	"unicode/utf8"
)

// maxSearchText é o tanto de texto de um artefato que entra na busca. Basta para achar um
// documento pelo que ele diz, e não deixa um HTML de 50 MB virar uma linha de 50 MB no banco.
const maxSearchText = 100_000

// invisibleElements são as tags cujo conteúdo não é texto que se lê na página.
var invisibleElements = []string{"script", "style", "noscript", "template", "svg"}

// htmlText tira as tags e devolve o texto que a pessoa lê na página, para a busca
// (content_text). Não é um parser de HTML completo — nem precisa: um texto um pouco sujo
// ainda acha o documento, e um parser seria mais uma dependência.
func htmlText(page string) string {
	var text strings.Builder
	for len(page) > 0 && text.Len() < maxSearchText {
		start := strings.IndexByte(page, '<')
		if start < 0 {
			text.WriteString(page)
			break
		}
		text.WriteString(page[:start])
		text.WriteByte(' ')
		page = skipTag(page[start:])
	}

	return cleanText(html.UnescapeString(text.String()))
}

// skipTag pula a tag que começa em page[0] — e, se ela abre um elemento invisível, pula
// também tudo até o fechamento dele.
func skipTag(page string) string {
	end := strings.IndexByte(page, '>')
	if end < 0 {
		return ""
	}
	tag := strings.ToLower(page[1:end])
	rest := page[end+1:]

	for _, element := range invisibleElements {
		if tag != element && !strings.HasPrefix(tag, element+" ") {
			continue
		}
		closing := strings.Index(strings.ToLower(rest), "</"+element)
		if closing < 0 {
			return ""
		}

		return rest[closing:]
	}

	return rest
}

// cleanText junta os espaços e corta no limite sem partir um caractere ao meio.
func cleanText(raw string) string {
	text := strings.Join(strings.Fields(raw), " ")
	if len(text) <= maxSearchText {
		return text
	}

	cut := maxSearchText
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}

	return text[:cut]
}
