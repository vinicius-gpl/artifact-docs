package artifacts

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/database"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/infra/storage"
)

// artifactsFolder é a pasta dos artefatos no bucket (briefing, "R2"). Em inglês: é
// infraestrutura.
const artifactsFolder = "artifacts/"

// sandboxPolicy é a proteção de servir HTML de terceiros no MESMO domínio do hub (briefing,
// "Segurança no mesmo domínio"). Sem allow-same-origin, o navegador trata o artefato como uma
// origem opaca: o script dele não lê o cookie nem chama a API como você.
const sandboxPolicy = "sandbox allow-scripts allow-popups allow-forms"

func artifactFolder(artifactID string) string {
	return artifactsFolder + artifactID + "/"
}

func versionFolder(artifactID string, version int) string {
	return fmt.Sprintf("%sv%d/", artifactFolder(artifactID), version)
}

// Serve é GET /a/{access}/{path...}: entrega um arquivo da versão no ar, direto do R2.
//
// O arquivo vai do R2 para a resposta aos pedaços (http.ServeContent), sem passar inteiro
// pela memória. O ServeContent também responde o 304 (ETag) e o pedido de faixa, que é como
// o navegador abre um PDF grande sem baixar tudo.
func (handlers *Handlers) Serve(writer http.ResponseWriter, request *http.Request) {
	access := handlers.linker.Read(request.PathValue("access"))
	served, err := handlers.resolve(request, access.LinkToken)
	if errors.Is(err, database.ErrNotFound) {
		http.Error(writer, "Artefato não encontrado.", http.StatusNotFound)
		return
	}
	if err != nil {
		handlers.logger.Error("serve artifact", "error", err)
		http.Error(writer, "Não consegui abrir o artefato. Tente de novo em instantes.", http.StatusInternalServerError)
		return
	}
	if served.IsPrivate && !access.Signed {
		http.Error(writer, "Este link é privado ou expirou. Abra o artefato pelo hub.", http.StatusForbidden)
		return
	}

	name, ok := servedPath(request.PathValue("path"), served.Format)
	if !ok {
		http.Error(writer, "Arquivo não encontrado.", http.StatusNotFound)
		return
	}

	if isEntryPage(request.PathValue("path"), served.Format) {
		handlers.streamEntryPage(writer, request, served, linkOf(request, access.LinkToken))
		return
	}

	handlers.stream(writer, request, served, name)
}

// isEntryPage diz se o pedido é a página que o link abre — a que ganha o cartão. Os assets e
// as subpáginas de um site vão como estão; o PDF não tem <head>.
func isEntryPage(requested string, format string) bool {
	return requested == "" && format != "pdf"
}

// RedirectToFolder é GET /a/{access}: manda para /a/{access}/. Sem a barra final, o
// navegador resolveria "style.css" para /a/style.css, fora da pasta do artefato.
func (handlers *Handlers) RedirectToFolder(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, request.URL.Path+"/", http.StatusMovedPermanently)
}

func (handlers *Handlers) resolve(request *http.Request, linkToken string) (Served, error) {
	if served, ok := handlers.served.get(linkToken); ok {
		return served, nil
	}

	served, err := handlers.store.FindServed(request.Context(), linkToken)
	if err != nil {
		return Served{}, err
	}
	handlers.served.put(linkToken, served)

	return served, nil
}

func (handlers *Handlers) stream(writer http.ResponseWriter, request *http.Request, served Served, name string) {
	object, ok := handlers.open(writer, request, served, name)
	if !ok {
		return
	}
	defer func() { _ = object.Content.Close() }()

	setServedHeaders(writer.Header(), served, name, object)
	http.ServeContent(writer, request, name, object.LastModified, object.Content)
}

// streamEntryPage serve o index.html com as tags do cartão do link no <head> (preview.go).
// A página inteira passa pela memória, então só até maxEntryPageBytes: acima disso ela vai
// como está, sem cartão.
func (handlers *Handlers) streamEntryPage(writer http.ResponseWriter, request *http.Request, served Served, link previewLink) {
	const name = "index.html"
	object, ok := handlers.open(writer, request, served, name)
	if !ok {
		return
	}
	defer func() { _ = object.Content.Close() }()

	setServedHeaders(writer.Header(), served, name, object)
	if object.Size > maxEntryPageBytes {
		http.ServeContent(writer, request, name, object.LastModified, object.Content)
		return
	}

	page, err := io.ReadAll(io.LimitReader(object.Content, maxEntryPageBytes))
	if err != nil {
		handlers.logger.Error("read artifact entry page", "artifact", served.ArtifactID, "error", err)
		http.Error(writer, "Não consegui abrir o arquivo. Tente de novo em instantes.", http.StatusBadGateway)
		return
	}

	page = injectPreview(page, served, link)
	writer.Header().Set("ETag", previewETag(object.ETag, page))
	http.ServeContent(writer, request, name, object.LastModified, bytes.NewReader(page))
}

// open busca o arquivo no R2 e já responde a falha: false é "a resposta foi dada".
func (handlers *Handlers) open(writer http.ResponseWriter, request *http.Request, served Served, name string) (storage.Object, bool) {
	object, err := handlers.files.Get(request.Context(), versionFolder(served.ArtifactID, served.Version)+name)
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(writer, "Arquivo não encontrado.", http.StatusNotFound)
		return storage.Object{}, false
	}
	if err != nil {
		handlers.logger.Error("read artifact file", "artifact", served.ArtifactID, "file", name, "error", err)
		http.Error(writer, "Não consegui abrir o arquivo. Tente de novo em instantes.", http.StatusBadGateway)
		return storage.Object{}, false
	}

	return object, true
}

// setServedHeaders põe os cabeçalhos de todo arquivo servido.
func setServedHeaders(header http.Header, served Served, name string, object storage.Object) {
	contentType := object.ContentType
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = contentTypeOf(name)
	}
	header.Set("Content-Type", contentType)
	header.Set("ETag", `"`+strings.Trim(object.ETag, `"`)+`"`)
	// O endereço é o mesmo em todas as versões: o navegador guarda, mas confere a cada
	// visita (o ETag faz a conferência custar um 304 vazio).
	header.Set("Cache-Control", "no-cache")
	if served.IsPrivate {
		header.Set("Cache-Control", "private, no-cache")
	}
	// O link carrega o segredo do artefato: um link externo dentro dele não pode levá-lo no
	// Referer para outro site.
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Content-Type-Options", "nosniff")

	// O PDF fica sem sandbox: o leitor de PDF do Chrome se recusa a abrir dentro de um. Um
	// PDF não roda o JavaScript da página, então não há o que isolar.
	if path.Ext(name) != ".pdf" {
		header.Set("Content-Security-Policy", sandboxPolicy)
	}
}

// servedPath traduz o caminho pedido no nome do arquivo na pasta da versão. Pasta pede o
// index.html dela; o PDF é sempre o index.pdf.
func servedPath(requested string, format string) (string, bool) {
	if format == "pdf" && requested == "" {
		return "index.pdf", true
	}
	if requested == "" || strings.HasSuffix(requested, "/") {
		requested += "index.html"
	}
	if !isSafePath(requested) || isHidden(requested) && requested != thumbFile {
		return "", false
	}

	return requested, true
}
