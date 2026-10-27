package artifacts

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"strings"
)

// Os limites do que se publica. Passou, é recusado com a mensagem dizendo qual foi.
const (
	// maxUploadBytes é o maior envio: o arquivo, a miniatura e os campos juntos.
	maxUploadBytes = 100 << 20
	// maxZipEntries e maxZipUnpacked barram o .zip-bomba: poucos KB que viram GB ao abrir.
	maxZipEntries  = 2_000
	maxZipUnpacked = 300 << 20
	// maxTextSource é o tanto de HTML ou Markdown lido para o texto da busca e a conversão.
	maxTextSource = 5 << 20
)

// invalidUpload é o erro de quem enviou — volta como 400, com a mensagem para a tela.
type invalidUpload struct {
	message string
}

func (failure invalidUpload) Error() string {
	return failure.message
}

// bundle é o que uma versão grava no R2: os arquivos, o formato e o texto para a busca.
type bundle struct {
	format string
	files  []bundleFile
	text   string
	// release fecha o que o bundle abriu (o .zip); o temporário é apagado por quem recebeu.
	release func() error
}

// bundleFile é um arquivo da versão, aberto só na hora de enviar ao R2.
type bundleFile struct {
	path        string
	size        int64
	contentType string
	open        func() (io.ReadCloser, error)
}

// upload é o arquivo recebido, já gravado num temporário em disco.
type upload struct {
	diskPath string
	name     string
	size     int64
}

// buildBundle decide pelo nome do arquivo o que publicar (briefing, módulo 1).
func buildBundle(received upload, title string) (bundle, error) {
	switch strings.ToLower(path.Ext(received.name)) {
	case ".html", ".htm":
		return htmlBundle(received)
	case ".zip":
		return zipBundle(received)
	case ".md", ".markdown":
		return markdownBundle(received, title)
	case ".pdf":
		return pdfBundle(received)
	default:
		return bundle{}, invalidUpload{"Envie um arquivo .html, .zip, .md ou .pdf."}
	}
}

func htmlBundle(received upload) (bundle, error) {
	source, err := readHead(received.diskPath)
	if err != nil {
		return bundle{}, err
	}

	return bundle{
		format:  "html",
		files:   []bundleFile{diskFile("index.html", received)},
		text:    htmlText(string(source)),
		release: noRelease,
	}, nil
}

func markdownBundle(received upload, title string) (bundle, error) {
	if received.size > maxTextSource {
		return bundle{}, invalidUpload{"O Markdown passa de 5 MB. Divida em mais de um documento."}
	}
	source, err := os.ReadFile(received.diskPath)
	if err != nil {
		return bundle{}, err
	}

	page, err := renderMarkdown(source, title)
	if err != nil {
		return bundle{}, err
	}

	return bundle{
		format:  "md",
		files:   []bundleFile{memoryFile("index.html", page)},
		text:    cleanText(string(source)),
		release: noRelease,
	}, nil
}

// pdfBundle confere a assinatura do arquivo: um .exe renomeado para .pdf não passa.
func pdfBundle(received upload) (bundle, error) {
	head, err := readHead(received.diskPath)
	if err != nil {
		return bundle{}, err
	}
	if !bytes.HasPrefix(head, []byte("%PDF-")) {
		return bundle{}, invalidUpload{"O arquivo não é um PDF válido."}
	}

	return bundle{
		format:  "pdf",
		files:   []bundleFile{diskFile("index.pdf", received)},
		release: noRelease,
	}, nil
}

// zipBundle abre o .zip de um site: index.html e os assets. O site pode vir na raiz do .zip
// ou dentro de uma pasta só (o que o "Compactar" do sistema costuma fazer).
func zipBundle(received upload) (bundle, error) {
	archive, err := zip.OpenReader(received.diskPath)
	if err != nil {
		return bundle{}, invalidUpload{"Não consegui abrir o .zip. Ele pode estar corrompido."}
	}

	entries, err := siteEntries(archive.File)
	if err != nil {
		_ = archive.Close()
		return bundle{}, err
	}

	root, err := siteRoot(entries)
	if err != nil {
		_ = archive.Close()
		return bundle{}, err
	}

	files := make([]bundleFile, 0, len(entries))
	var text string
	for _, entry := range entries {
		sitePath := strings.TrimPrefix(entry.Name, root)
		files = append(files, zipFile(sitePath, entry))
		if sitePath == "index.html" {
			text, err = zipEntryText(entry)
		}
		if err != nil {
			_ = archive.Close()
			return bundle{}, err
		}
	}

	return bundle{format: "html", files: files, text: text, release: archive.Close}, nil
}

// siteEntries filtra o .zip: tira pastas e arquivos ocultos (.DS_Store, __MACOSX), recusa
// caminho perigoso e confere os limites.
func siteEntries(all []*zip.File) ([]*zip.File, error) {
	var entries []*zip.File
	var unpacked uint64
	for _, entry := range all {
		// O caminho perigoso é conferido ANTES de "oculto": `../x` começa com ponto, e na ordem
		// inversa um .zip malicioso passava calado, com a entrada descartada em vez de recusada.
		// Pasta vem com barra no fim (`assets/`), que o Clean tiraria; ela não muda a segurança.
		if !isSafePath(strings.TrimSuffix(entry.Name, "/")) {
			return nil, invalidUpload{fmt.Sprintf("O .zip tem um caminho inválido: %s", entry.Name)}
		}
		if entry.FileInfo().IsDir() || isHidden(entry.Name) {
			continue
		}

		unpacked += entry.UncompressedSize64
		entries = append(entries, entry)
	}

	if len(entries) > maxZipEntries {
		return nil, invalidUpload{fmt.Sprintf("O .zip tem mais de %d arquivos.", maxZipEntries)}
	}
	if unpacked > maxZipUnpacked {
		return nil, invalidUpload{"O .zip passa de 300 MB descompactado."}
	}

	return entries, nil
}

// siteRoot acha onde está o index.html: na raiz, ou numa pasta única que contém tudo.
func siteRoot(entries []*zip.File) (string, error) {
	notFound := invalidUpload{"O .zip precisa ter um index.html na raiz (ou numa pasta só)."}
	if len(entries) == 0 {
		return "", notFound
	}

	folder, _, nested := strings.Cut(entries[0].Name, "/")
	prefix := folder + "/"
	hasRootIndex, hasFolderIndex, allInFolder := false, false, nested
	for _, entry := range entries {
		hasRootIndex = hasRootIndex || entry.Name == "index.html"
		hasFolderIndex = hasFolderIndex || entry.Name == prefix+"index.html"
		allInFolder = allInFolder && strings.HasPrefix(entry.Name, prefix)
	}

	if hasRootIndex {
		return "", nil
	}
	if allInFolder && hasFolderIndex {
		return prefix, nil
	}

	return "", notFound
}

// isSafePath recusa o que sairia da pasta da versão no R2: "..", caminho absoluto, barra
// invertida do Windows, caminho que o Clean mudaria (a/./b, a//b).
func isSafePath(name string) bool {
	if name == "" || strings.HasPrefix(name, "/") || strings.Contains(name, `\`) {
		return false
	}
	if path.Clean(name) != name {
		return false
	}

	for _, segment := range strings.Split(name, "/") {
		if segment == ".." {
			return false
		}
	}

	return true
}

// isHidden é o arquivo que o sistema põe no .zip sem ninguém pedir: .DS_Store, __MACOSX/.
// Tirá-los também garante que nenhum arquivo do site use o nome reservado da miniatura.
func isHidden(name string) bool {
	for _, segment := range strings.Split(name, "/") {
		if strings.HasPrefix(segment, ".") || segment == "__MACOSX" {
			return true
		}
	}

	return false
}

func zipEntryText(entry *zip.File) (string, error) {
	content, err := entry.Open()
	if err != nil {
		return "", invalidUpload{"Não consegui ler o index.html do .zip."}
	}
	defer func() { _ = content.Close() }()

	source, err := io.ReadAll(io.LimitReader(content, maxTextSource))
	if err != nil {
		return "", invalidUpload{"Não consegui ler o index.html do .zip."}
	}

	return htmlText(string(source)), nil
}

// readHead lê o começo do arquivo: o bastante para o texto da busca ou a assinatura do PDF.
func readHead(diskPath string) ([]byte, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	return io.ReadAll(io.LimitReader(file, maxTextSource))
}

func diskFile(sitePath string, received upload) bundleFile {
	return bundleFile{
		path:        sitePath,
		size:        received.size,
		contentType: contentTypeOf(sitePath),
		open:        func() (io.ReadCloser, error) { return os.Open(received.diskPath) },
	}
}

func memoryFile(sitePath string, content []byte) bundleFile {
	return bundleFile{
		path:        sitePath,
		size:        int64(len(content)),
		contentType: contentTypeOf(sitePath),
		open:        func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(content)), nil },
	}
}

func zipFile(sitePath string, entry *zip.File) bundleFile {
	return bundleFile{
		path:        sitePath,
		size:        int64(entry.UncompressedSize64),
		contentType: contentTypeOf(sitePath),
		open:        entry.Open,
	}
}

// contentTypeOf decide o tipo pela extensão. O que o Go não conhece vai como binário: o
// navegador baixa em vez de tentar adivinhar (e o nosniff impede que ele adivinhe).
func contentTypeOf(sitePath string) string {
	if contentType := mime.TypeByExtension(path.Ext(sitePath)); contentType != "" {
		return contentType
	}

	return "application/octet-stream"
}

func noRelease() error {
	return nil
}
