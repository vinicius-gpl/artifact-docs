package artifacts

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"html"
	"net/http"
	"strings"
)

// O PREVIEW DO LINK: quem cola /a/<token>/ no WhatsApp, Telegram, Slack ou Discord vê um cartão
// com título, descrição e imagem. Esses apps não rodam o JavaScript da página: leem as tags
// Open Graph do <head>, e o WhatsApp só lê o começo do HTML. Por isso as tags entram logo
// depois do <head>, na hora de servir, com o título e a descrição que estão no banco agora —
// editar o artefato muda o cartão sem subir versão nova.
//
// Artefato PRIVADO não ganha cartão: o preview mostraria o título e a descrição a quem só
// recebeu a mensagem. Toda página de entrada ganha o `noindex`: o hub é pessoal, e um link
// compartilhado não deve virar resultado de busca.

// maxEntryPageBytes é o maior index.html que passa pela injeção. Acima disso a página vai
// como está, sem cartão: carregá-la inteira na memória custaria mais que o preview vale.
const maxEntryPageBytes = 2 << 20

// defaultCardPath é a imagem do cartão quando a versão não tem miniatura: o cartão padrão da
// marca, servido pela SPA (svelte/static/og-card.png).
const defaultCardPath = "/og-card.png"

// defaultCardWidth e defaultCardHeight são as medidas do cartão padrão (1.91:1, o formato que
// os apps esperam para a imagem grande).
const (
	defaultCardWidth  = "1200"
	defaultCardHeight = "630"
)

const siteName = "artifact-docs"

// kindDescriptions é a descrição do cartão quando o artefato não tem uma.
var kindDescriptions = map[string]string{
	"document": "Documento no artifact-docs",
	"plan":     "Planejamento no artifact-docs",
	"idea":     "Ideia no artifact-docs",
}

// previewLink é o que o cartão precisa saber do pedido: o endereço absoluto da página.
type previewLink struct {
	origin    string
	linkToken string
}

// linkOf monta o endereço absoluto a partir do próprio pedido. O Caddy repassa o Host e o
// X-Forwarded-Proto do visitante: assim o hub não precisa de mais uma variável com o próprio
// domínio, e o endereço continua certo em localhost.
func linkOf(request *http.Request, linkToken string) previewLink {
	scheme := "http"
	if request.TLS != nil {
		scheme = "https"
	}
	forwarded := request.Header.Get("X-Forwarded-Proto")
	if forwarded == "https" || forwarded == "http" {
		scheme = forwarded
	}

	return previewLink{origin: scheme + "://" + request.Host, linkToken: linkToken}
}

// previewTags são as tags que entram no <head>. `hasTitle` diz se a página já tem <title>: só
// sem ele o cartão põe o do artefato, para não brigar com o que a página escolheu.
func previewTags(served Served, link previewLink, hasTitle bool) string {
	var tags strings.Builder
	tags.WriteString(`<meta name="robots" content="noindex">`)
	if served.IsPrivate {
		return tags.String()
	}

	pageURL := link.origin + "/a/" + link.linkToken + "/"
	writeMeta(&tags, "og:site_name", siteName)
	writeMeta(&tags, "og:type", "article")
	writeMeta(&tags, "og:title", served.Title)
	writeMeta(&tags, "og:description", descriptionOf(served))
	writeMeta(&tags, "og:url", pageURL)
	writeImage(&tags, served, link, pageURL)
	tags.WriteString(`<meta name="twitter:card" content="summary_large_image">`)
	if !hasTitle {
		tags.WriteString("<title>" + html.EscapeString(served.Title) + "</title>")
	}

	return tags.String()
}

// writeImage escolhe a imagem: a miniatura da versão, quando quem publicou mandou uma; senão,
// o cartão padrão da marca, que tem medida conhecida.
func writeImage(tags *strings.Builder, served Served, link previewLink, pageURL string) {
	if served.HasThumb {
		writeMeta(tags, "og:image", pageURL+thumbFile)
		writeMeta(tags, "og:image:type", "image/png")
		return
	}

	writeMeta(tags, "og:image", link.origin+defaultCardPath)
	writeMeta(tags, "og:image:type", "image/png")
	writeMeta(tags, "og:image:width", defaultCardWidth)
	writeMeta(tags, "og:image:height", defaultCardHeight)
}

func writeMeta(tags *strings.Builder, property, content string) {
	tags.WriteString(`<meta property="` + property + `" content="` + html.EscapeString(content) + `">`)
}

func descriptionOf(served Served) string {
	if served.Description != nil && strings.TrimSpace(*served.Description) != "" {
		return strings.TrimSpace(*served.Description)
	}

	return kindDescriptions[served.Kind]
}

// injectPreview põe as tags logo depois da abertura do <head>. Página sem <head> ganha um —
// depois do <html>, se houver, ou no começo. A busca é sem diferença de maiúscula: HTML
// escrito à mão chega com <HEAD> também.
func injectPreview(page []byte, served Served, link previewLink) []byte {
	lower := bytes.ToLower(page)
	tags := previewTags(served, link, bytes.Contains(lower, []byte("<title")))

	if at := tagEnd(lower, "<head"); at >= 0 {
		return insertAt(page, at, tags)
	}
	if at := tagEnd(lower, "<html"); at >= 0 {
		return insertAt(page, at, "<head>"+tags+"</head>")
	}

	return insertAt(page, 0, "<head>"+tags+"</head>")
}

// tagEnd devolve a posição logo depois do `>` da primeira tag `name` (ex.: "<head"), ou -1.
// Exige que o nome termine ali (`<header>` não é `<head>`).
func tagEnd(lower []byte, name string) int {
	offset := 0
	for {
		found := bytes.Index(lower[offset:], []byte(name))
		if found < 0 {
			return -1
		}
		start := offset + found + len(name)
		if start < len(lower) && isTagNameEnd(lower[start]) {
			closing := bytes.IndexByte(lower[start:], '>')
			if closing < 0 {
				return -1
			}
			return start + closing + 1
		}
		offset = start
	}
}

func isTagNameEnd(char byte) bool {
	return char == '>' || char == ' ' || char == '\t' || char == '\n' || char == '\r' || char == '/'
}

func insertAt(page []byte, at int, text string) []byte {
	result := make([]byte, 0, len(page)+len(text))
	result = append(result, page[:at]...)
	result = append(result, text...)

	return append(result, page[at:]...)
}

// previewETag muda junto com o cartão: o mesmo arquivo com outro título (ou virado privado)
// não pode responder 304 com o cartão velho.
func previewETag(objectETag string, page []byte) string {
	sum := sha256.Sum256(page)

	return `"` + strings.Trim(objectETag, `"`) + "-" + hex.EncodeToString(sum[:6]) + `"`
}
