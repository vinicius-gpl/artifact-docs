package apps

import (
	"context"
	"html"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// O ícone de cada app sai da página do próprio app: o <link rel="icon"> que o navegador usa na
// aba, ou o /favicon.ico de sempre. Só o ENDEREÇO é guardado — a tela carrega a imagem direto
// do site, sem passar pela VPS nem pelo R2.

const (
	// pageReadLimit é quanto da página é lido. Os <link> moram no <head>, que cabe com folga
	// nisso; ler a página inteira de um app pesado gastaria memória da VPS à toa.
	pageReadLimit = 256 << 10
	// monitorUserAgent se parece com navegador porque há proxy (Cloudflare) que devolve página
	// de bloqueio, sem os <link>, para quem não se apresenta.
	monitorUserAgent = "Mozilla/5.0 (compatible; artifact-docs-monitor)"
)

// A preferência entre os ícones da página: SVG fica nítido em qualquer tamanho; o
// apple-touch-icon costuma ter 180 px; o resto ganha pelo maior `sizes`.
const (
	scoreSVG        = 3000
	scoreAppleTouch = 2000
	scoreIcon       = 1000
	maxSizeBonus    = 999
)

var (
	linkTagPattern   = regexp.MustCompile(`(?is)<link\b[^>]*>`)
	attributePattern = regexp.MustCompile(`([a-zA-Z-]+)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'>]+))`)
)

// discoverIcon devolve o endereço absoluto do ícone do app, ou "" quando não achou nenhum —
// aí a tela fica com as iniciais.
func discoverIcon(ctx context.Context, client *http.Client, pageURL string) string {
	if iconURL := iconFromPage(ctx, client, pageURL); iconURL != "" {
		return iconURL
	}

	return fallbackFavicon(ctx, client, pageURL)
}

// iconFromPage lê o começo da página e escolhe o melhor <link rel="icon">. Os endereços
// relativos são resolvidos contra o endereço FINAL, depois dos redirecionamentos: é dele que o
// navegador resolveria.
func iconFromPage(ctx context.Context, client *http.Client, pageURL string) string {
	response, err := fetch(ctx, client, pageURL)
	if err != nil {
		return ""
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode >= http.StatusBadRequest {
		return ""
	}

	document, err := io.ReadAll(io.LimitReader(response.Body, pageReadLimit))
	if err != nil {
		return ""
	}

	return bestIcon(string(document), response.Request.URL)
}

// fallbackFavicon tenta o /favicon.ico da origem — o lugar que o navegador procura quando a
// página não diz nada. Só vale se a resposta for mesmo uma imagem: muito site responde 200 com
// a própria página para qualquer caminho.
func fallbackFavicon(ctx context.Context, client *http.Client, pageURL string) string {
	page, err := url.Parse(pageURL)
	if err != nil || page.Host == "" {
		return ""
	}
	faviconURL := (&url.URL{Scheme: page.Scheme, Host: page.Host, Path: "/favicon.ico"}).String()

	response, err := fetch(ctx, client, faviconURL)
	if err != nil {
		return ""
	}
	_ = response.Body.Close()

	isImage := strings.HasPrefix(response.Header.Get("Content-Type"), "image/")
	if response.StatusCode != http.StatusOK || !isImage {
		return ""
	}

	return faviconURL
}

func fetch(ctx context.Context, client *http.Client, address string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", monitorUserAgent)

	return client.Do(request)
}

// bestIcon percorre os <link> do documento e fica com o de maior nota; empate fica com o
// primeiro, que é a ordem que o autor da página escolheu.
func bestIcon(document string, base *url.URL) string {
	best, bestScore := "", 0
	for _, tag := range linkTagPattern.FindAllString(document, -1) {
		attributes := attributesOf(tag)
		score := iconScore(attributes)
		if score <= bestScore {
			continue
		}

		resolved := resolveIcon(base, attributes["href"])
		if resolved == "" {
			continue
		}
		best, bestScore = resolved, score
	}

	return best
}

// attributesOf lê os atributos de uma tag, com o nome em minúsculas e as entidades (&amp;)
// já desfeitas.
func attributesOf(tag string) map[string]string {
	attributes := map[string]string{}
	for _, match := range attributePattern.FindAllStringSubmatch(tag, -1) {
		value := match[2] + match[3] + match[4]
		attributes[strings.ToLower(match[1])] = html.UnescapeString(value)
	}

	return attributes
}

// iconScore dá nota a um <link>: zero quando ele não é ícone (stylesheet, preload…).
func iconScore(attributes map[string]string) int {
	relations := strings.Fields(strings.ToLower(attributes["rel"]))
	isAppleTouch := slices.Contains(relations, "apple-touch-icon") || slices.Contains(relations, "apple-touch-icon-precomposed")
	if !slices.Contains(relations, "icon") && !isAppleTouch {
		return 0
	}
	if isSVGIcon(attributes) {
		return scoreSVG
	}
	if isAppleTouch {
		return scoreAppleTouch
	}

	return scoreIcon + largestSize(attributes["sizes"])
}

func isSVGIcon(attributes map[string]string) bool {
	if strings.EqualFold(attributes["type"], "image/svg+xml") {
		return true
	}

	return strings.EqualFold(path.Ext(strings.SplitN(attributes["href"], "?", 2)[0]), ".svg")
}

// largestSize lê o `sizes` ("16x16 32x32", ou "any") e devolve o maior lado, com teto para não
// passar à frente do SVG.
func largestSize(sizes string) int {
	largest := 0
	for _, size := range strings.Fields(strings.ToLower(sizes)) {
		if size == "any" {
			return maxSizeBonus
		}
		width, err := strconv.Atoi(strings.SplitN(size, "x", 2)[0])
		if err == nil && width > largest {
			largest = width
		}
	}

	return min(largest, maxSizeBonus)
}

// resolveIcon transforma o href em endereço absoluto. Só http e https passam: um `data:` iria
// inteiro para o banco, e um `javascript:` nunca é imagem.
func resolveIcon(base *url.URL, href string) string {
	reference, err := url.Parse(strings.TrimSpace(href))
	if err != nil || href == "" {
		return ""
	}

	resolved := base.ResolveReference(reference)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}

	return resolved.String()
}
