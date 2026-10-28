package artifacts

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/httpjson"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/modules/auth"
)

// Publicar é PUT /api/artifacts/{label}, em multipart: o arquivo e, opcionalmente, a
// miniatura e os campos. É a mesma rota para a tela e para as skills (briefing, módulo 6):
//
//	curl -X PUT -H "Authorization: Bearer …" -F file=@site.zip -F title="Guia" \
//	     -F kind=document https://hub/api/artifacts/guia
//
// Cada envio é uma versão nova. Os arquivos sobem para o R2 ANTES de o banco mudar: a versão
// só aparece quando está inteira lá, e uma falha no meio deixa a anterior no ar.

const (
	// maxThumbBytes é a maior miniatura aceita. Um PNG de card raramente passa de 300 KB.
	maxThumbBytes = 2 << 20
	// maxFieldBytes é o maior campo de texto (título, descrição, tags).
	maxFieldBytes = 4 << 10
	maxTitle      = 200
	maxTags       = 20
	maxTagLength  = 40
	// cleanupTimeout é quanto a limpeza de uma publicação que falhou tem, já fora do pedido
	// (que pode ter sido cancelado).
	cleanupTimeout = 30 * time.Second
)

var (
	labelPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	pngSignature = []byte("\x89PNG\r\n\x1a\n")
)

// publishForm é o que veio no multipart.
type publishForm struct {
	meta     PublishMeta
	received upload
	hasFile  bool
	thumb    []byte
}

// Publish é PUT /api/artifacts/{label} → artifactSchema (201 quando o artefato é novo).
func (handlers *Handlers) Publish(writer http.ResponseWriter, request *http.Request) {
	label := request.PathValue("label")
	if len(label) > 80 || !labelPattern.MatchString(label) {
		httpjson.Error(writer, http.StatusBadRequest, "Use só letras minúsculas, números e hífen na label (até 80).")
		return
	}

	form, err := readPublishForm(writer, request)
	defer removeUpload(form.received)
	if handlers.answeredPublishFailure(writer, request, err) {
		return
	}

	artifact, created, err := handlers.publishVersion(request.Context(), auth.CurrentUser(request).ID, label, form)
	if handlers.answeredPublishFailure(writer, request, err) {
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	httpjson.Write(writer, status, handlers.linker.withURLs(artifact))
}

func (handlers *Handlers) answeredPublishFailure(writer http.ResponseWriter, request *http.Request, err error) bool {
	if err == nil {
		return false
	}

	var invalid invalidUpload
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &invalid):
		httpjson.Error(writer, http.StatusBadRequest, invalid.message)
	case errors.As(err, &tooLarge):
		httpjson.Error(writer, http.StatusRequestEntityTooLarge, "O envio passa de 100 MB.")
	case errors.Is(err, ErrConflict):
		httpjson.Error(writer, http.StatusConflict, "Outra publicação desta label terminou antes. Envie de novo.")
	default:
		httpjson.Fail(handlers.logger, writer, request, err, "Não consegui publicar. Tente de novo em instantes.")
	}

	return true
}

// publishVersion sobe os arquivos e grava a versão. Devolve o artefato e se ele é novo.
func (handlers *Handlers) publishVersion(ctx context.Context, ownerID, label string, form publishForm) (Artifact, bool, error) {
	target, err := handlers.store.PublishTarget(ctx, ownerID, label)
	if err != nil {
		return Artifact{}, false, err
	}
	if !target.Exists {
		if err := requireNewFields(form.meta); err != nil {
			return Artifact{}, false, err
		}
		target.ArtifactID = newUUID()
		applyNewDefaults(&form.meta)
	}

	content, err := buildBundle(form.received, pageTitle(form.meta, label))
	if err != nil {
		return Artifact{}, false, err
	}
	defer func() { _ = content.release() }()

	version := target.LatestVersion + 1
	folder := versionFolder(target.ArtifactID, version)
	record := VersionRecord{
		IsNew: !target.Exists, ArtifactID: target.ArtifactID, OwnerID: ownerID, Label: label,
		Meta: form.meta, Format: content.format, ContentText: content.text, Version: version,
		HasThumb: len(form.thumb) > 0,
	}

	record.SizeBytes, err = handlers.uploadVersion(ctx, folder, content.files, form.thumb)
	if err == nil {
		err = handlers.store.SaveVersion(ctx, record)
	}
	if err != nil {
		handlers.discard(folder)
		return Artifact{}, false, err
	}

	handlers.served.forget(target.LinkToken)
	artifact, err := handlers.store.Find(ctx, ownerID, label)

	return artifact, !target.Exists, err
}

// uploadVersion envia os arquivos e a miniatura para a pasta da versão e devolve o tamanho
// somado (a miniatura não conta: ela não é do artefato).
func (handlers *Handlers) uploadVersion(ctx context.Context, folder string, files []bundleFile, thumb []byte) (int64, error) {
	var total int64
	for _, file := range files {
		if err := handlers.uploadFile(ctx, folder, file); err != nil {
			return 0, err
		}
		total += file.size
	}

	if len(thumb) == 0 {
		return total, nil
	}

	return total, handlers.files.Put(ctx, folder+thumbFile, bytes.NewReader(thumb), int64(len(thumb)), "image/png")
}

func (handlers *Handlers) uploadFile(ctx context.Context, folder string, file bundleFile) error {
	content, err := file.open()
	if err != nil {
		return err
	}
	defer func() { _ = content.Close() }()

	return handlers.files.Put(ctx, folder+file.path, content, file.size, file.contentType)
}

// discard apaga o que uma publicação que falhou chegou a subir. Se a limpeza também falhar,
// sobra só arquivo numa pasta que nenhum ponteiro aponta: não aparece em lugar nenhum.
func (handlers *Handlers) discard(folder string) {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	if err := handlers.files.DeleteFolder(ctx, folder); err != nil {
		handlers.logger.Error("failed publish left files behind", "folder", folder, "error", err)
	}
}

func requireNewFields(meta PublishMeta) error {
	if meta.Title == nil || meta.Kind == nil {
		return invalidUpload{"Artefato novo precisa de título e tipo (title e kind)."}
	}

	return nil
}

// applyNewDefaults completa o artefato novo. Privado por padrão: o hub é de uma pessoa só, e
// publicar aberto precisa ser uma escolha, não um esquecimento. Planejamento e ideia nascem
// como rascunho; documento não tem situação, e uma enviada por engano é descartada.
func applyNewDefaults(meta *PublishMeta) {
	if meta.IsPrivate == nil {
		private := true
		meta.IsPrivate = &private
	}
	if *meta.Kind == "document" {
		meta.Status = nil
		return
	}
	if meta.Status == nil {
		draft := "draft"
		meta.Status = &draft
	}
}

// pageTitle é o título da página gerada do Markdown: o que veio, ou a label.
func pageTitle(meta PublishMeta, label string) string {
	if meta.Title != nil {
		return *meta.Title
	}

	return label
}

// readPublishForm lê o multipart aos pedaços: o arquivo vai direto para um temporário em
// disco, sem passar inteiro pela memória.
func readPublishForm(writer http.ResponseWriter, request *http.Request) (publishForm, error) {
	request.Body = http.MaxBytesReader(writer, request.Body, maxUploadBytes)
	reader, err := request.MultipartReader()
	if err != nil {
		return publishForm{}, invalidUpload{"Envie o arquivo em multipart/form-data, no campo file."}
	}

	var form publishForm
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return form, err
		}
		err = readPart(part, &form)
		_ = part.Close()
		if err != nil {
			return form, err
		}
	}

	if !form.hasFile {
		return form, invalidUpload{"Envie o arquivo no campo file."}
	}

	return form, nil
}

func readPart(part *multipart.Part, form *publishForm) error {
	switch part.FormName() {
	case "file":
		received, err := saveUpload(part)
		form.received, form.hasFile = received, err == nil

		return err
	case "thumb":
		thumb, err := readThumb(part)
		form.thumb = thumb

		return err
	default:
		return readField(part, &form.meta)
	}
}

func saveUpload(part *multipart.Part) (upload, error) {
	name := path.Base(strings.ReplaceAll(part.FileName(), `\`, "/"))
	temporary, err := os.CreateTemp("", "artifact-upload-*")
	if err != nil {
		return upload{}, err
	}
	defer func() { _ = temporary.Close() }()

	size, err := io.Copy(temporary, part)
	received := upload{diskPath: temporary.Name(), name: name, size: size}
	if err != nil {
		return received, err
	}
	if size == 0 {
		return received, invalidUpload{"O arquivo enviado está vazio."}
	}

	return received, nil
}

func readThumb(part *multipart.Part) ([]byte, error) {
	thumb, err := io.ReadAll(io.LimitReader(part, maxThumbBytes+1))
	if err != nil {
		return nil, err
	}
	if len(thumb) > maxThumbBytes {
		return nil, invalidUpload{"A miniatura passa de 2 MB."}
	}
	if len(thumb) > 0 && !bytes.HasPrefix(thumb, pngSignature) {
		return nil, invalidUpload{"A miniatura precisa ser PNG."}
	}

	return thumb, nil
}

// readField lê um campo de texto. Campo que não é do formulário é lido e ignorado.
func readField(part *multipart.Part, meta *PublishMeta) error {
	raw, err := io.ReadAll(io.LimitReader(part, maxFieldBytes+1))
	if err != nil {
		return err
	}
	if len(raw) > maxFieldBytes {
		return invalidUpload{fmt.Sprintf("O campo %s passa de 4 KB.", part.FormName())}
	}
	value := strings.TrimSpace(string(raw))

	switch part.FormName() {
	case "title":
		return setTitle(meta, value)
	case "kind":
		return setKind(meta, value)
	case "description":
		meta.Description = &value
	case "tags":
		return setTags(meta, value)
	case "status":
		return setStatus(meta, value)
	case "private":
		return setPrivate(meta, value)
	}

	return nil
}

func setTitle(meta *PublishMeta, value string) error {
	if err := checkTitle(value); err != nil {
		return err
	}
	meta.Title = &value

	return nil
}

// checkTitle é a regra do título, a mesma ao publicar e ao editar (PATCH).
func checkTitle(value string) error {
	if value == "" || len([]rune(value)) > maxTitle {
		return invalidUpload{fmt.Sprintf("O título precisa ter de 1 a %d caracteres.", maxTitle)}
	}

	return nil
}

func setKind(meta *PublishMeta, value string) error {
	if !knownKinds[value] {
		return invalidUpload{"O tipo (kind) precisa ser document, plan ou idea."}
	}
	meta.Kind = &value

	return nil
}

func setStatus(meta *PublishMeta, value string) error {
	if !knownStatuses[value] {
		return invalidUpload{"A situação (status) precisa ser draft, active, done ou archived."}
	}
	meta.Status = &value

	return nil
}

func setPrivate(meta *PublishMeta, value string) error {
	if value != "true" && value != "false" {
		return invalidUpload{"O campo private precisa ser true ou false."}
	}
	private := value == "true"
	meta.IsPrivate = &private

	return nil
}

// setTags lê as tags separadas por vírgula, sem repetir. Vazio limpa as tags.
func setTags(meta *PublishMeta, value string) error {
	tags, err := normalizeTags(strings.Split(value, ","))
	if err != nil {
		return err
	}
	meta.Tags = tags

	return nil
}

// normalizeTags apara, tira as vazias e as repetidas e confere os limites — a mesma regra ao
// publicar (texto com vírgulas) e ao editar (lista do PATCH). Nunca devolve nil: lista vazia
// é "sem tags", e não "mantém".
func normalizeTags(raw []string) ([]string, error) {
	tags := []string{}
	seen := map[string]bool{}
	for _, tag := range raw {
		tag = strings.TrimSpace(tag)
		if tag == "" || seen[tag] {
			continue
		}
		if len([]rune(tag)) > maxTagLength {
			return nil, invalidUpload{fmt.Sprintf("Cada tag tem no máximo %d caracteres.", maxTagLength)}
		}
		seen[tag] = true
		tags = append(tags, tag)
	}
	if len(tags) > maxTags {
		return nil, invalidUpload{fmt.Sprintf("Use no máximo %d tags.", maxTags)}
	}

	return tags, nil
}

func removeUpload(received upload) {
	if received.diskPath != "" {
		_ = os.Remove(received.diskPath)
	}
}

// newUUID é um UUID v4 — o id do artefato novo, sorteado aqui porque os arquivos sobem para
// artifacts/<id>/ antes de a linha existir no banco.
func newUUID() string {
	var raw [16]byte
	_, _ = rand.Read(raw[:])
	raw[6] = raw[6]&0x0f | 0x40
	raw[8] = raw[8]&0x3f | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", raw[0:4], raw[4:6], raw[6:8], raw[8:10], raw[10:16])
}
