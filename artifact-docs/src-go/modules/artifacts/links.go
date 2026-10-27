package artifacts

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

// Os links servidos (/a/<acesso>/) moram aqui (briefing, "Link do artefato").
//
// O <acesso> é o link_token do artefato. Para um PRIVADO, o token sozinho não abre: o link
// leva também um prazo e uma assinatura, no próprio caminho — /a/<token>~<prazo>~<assinatura>/.
// No caminho, e não na query string, para os assets relativos do HTML (style.css, img/…)
// herdarem a assinatura sem ninguém reescrever o HTML.
//
// O link privado é pedido pela tela à API autenticada a cada lista; o <iframe> que o abre
// não manda o token de login, e nem precisa.

// linkWindow é o tamanho da janela de validade. O prazo é arredondado para a janela: o mesmo
// artefato gera o MESMO link por horas, e o <iframe> da galeria não recarrega a cada
// atualização da lista. Um link vale entre uma e duas janelas.
const linkWindow = 6 * time.Hour

// signatureLength é quantos caracteres da assinatura vão no link: 22 em base64 são 132 bits.
const signatureLength = 22

const accessSeparator = "~"

// Linker monta e confere os links servidos.
type Linker struct {
	secret []byte
	now    func() time.Time
}

// NewLinker assina com secret (ARTIFACT_LINK_SECRET).
func NewLinker(secret []byte) *Linker {
	return &Linker{secret: secret, now: time.Now}
}

// URL é o endereço servido de um artefato, assinado quando ele é privado.
func (linker *Linker) URL(linkToken string, isPrivate bool) string {
	if !isPrivate {
		return "/a/" + linkToken + "/"
	}

	expires := linker.now().Truncate(linkWindow).Add(2 * linkWindow).Unix()
	access := strings.Join([]string{linkToken, strconv.FormatInt(expires, 10), linker.sign(linkToken, expires)}, accessSeparator)

	return "/a/" + access + "/"
}

// Access é o que o segmento <acesso> de /a/<acesso>/ diz.
type Access struct {
	LinkToken string
	// Signed é true quando o link trouxe assinatura válida e dentro do prazo — é o que abre
	// um artefato privado.
	Signed bool
}

// Read separa o token e confere a assinatura. Assinatura errada ou vencida não é erro: o
// link vale como público, e quem decide se isso basta é quem sabe se o artefato é privado.
func (linker *Linker) Read(access string) Access {
	parts := strings.Split(access, accessSeparator)
	if len(parts) != 3 {
		return Access{LinkToken: parts[0]}
	}

	expires, err := strconv.ParseInt(parts[1], 10, 64)
	valid := err == nil &&
		linker.now().Unix() < expires &&
		hmac.Equal([]byte(parts[2]), []byte(linker.sign(parts[0], expires)))

	return Access{LinkToken: parts[0], Signed: valid}
}

func (linker *Linker) sign(linkToken string, expires int64) string {
	mac := hmac.New(sha256.New, linker.secret)
	mac.Write([]byte(linkToken + accessSeparator + strconv.FormatInt(expires, 10)))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))[:signatureLength]
}

// withURLs põe no artefato os endereços servidos. A tela nunca monta o link sozinha: ela
// recebe pronto.
func (linker *Linker) withURLs(artifact Artifact) Artifact {
	artifact.URL = linker.URL(artifact.linkToken, artifact.IsPrivate)
	if artifact.hasThumb {
		thumb := artifact.URL + thumbFile
		artifact.ThumbURL = &thumb
	}

	return artifact
}
