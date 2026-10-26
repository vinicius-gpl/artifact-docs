package auth

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// ErrInvalidToken é todo token que não prova nada: malformado, assinatura errada, algoritmo
// que não aceitamos ou vencido. Quem chama não precisa saber qual — para a tela é sempre
// "entre de novo", e detalhar o motivo só ajudaria quem tenta forjar.
var ErrInvalidToken = errors.New("auth: invalid token")

// User é quem o token diz que é. Os campos são os que o Neon Auth (Better Auth) põe no JWT.
type User struct {
	ID    string
	Email string
	Name  string
	Image *string
}

// KeyFinder acha a chave pública pelo kid — o KeySet em produção, um mapa nos testes.
type KeyFinder interface {
	Key(ctx context.Context, kid string) (crypto.PublicKey, error)
}

// Verifier confere o JWT do Neon Auth.
type Verifier struct {
	keys KeyFinder
	now  func() time.Time
}

// NewVerifier confere os tokens com as chaves de keys.
func NewVerifier(keys KeyFinder) *Verifier {
	return &Verifier{keys: keys, now: time.Now}
}

type header struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type claims struct {
	Sub   string  `json:"sub"`
	Email string  `json:"email"`
	Name  string  `json:"name"`
	Image *string `json:"image"`
	Exp   int64   `json:"exp"`
}

// Verify devolve o dono do token, ou ErrInvalidToken.
//
// A ordem importa: a assinatura é conferida ANTES de ler as claims. Ler primeiro seria
// acreditar no que o token diz antes de saber se foi a Neon que disse.
func (verifier *Verifier) Verify(ctx context.Context, token string) (User, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return User{}, ErrInvalidToken
	}

	var head header
	if err := decodeSegment(parts[0], &head); err != nil {
		return User{}, ErrInvalidToken
	}

	key, err := verifier.keys.Key(ctx, head.Kid)
	if err != nil {
		return User{}, ErrInvalidToken
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !verifySignature(head.Alg, key, parts[0]+"."+parts[1], signature) {
		return User{}, ErrInvalidToken
	}

	var body claims
	if err := decodeSegment(parts[1], &body); err != nil {
		return User{}, ErrInvalidToken
	}

	return verifier.userFrom(body)
}

func (verifier *Verifier) userFrom(body claims) (User, error) {
	if body.Sub == "" || body.Email == "" || body.Exp == 0 {
		return User{}, ErrInvalidToken
	}
	if !verifier.now().Before(time.Unix(body.Exp, 0)) {
		return User{}, ErrInvalidToken
	}

	// O nome é opcional no cadastro; a tela precisa de algo para mostrar no avatar.
	name := body.Name
	if name == "" {
		name = body.Email
	}

	return User{ID: body.Sub, Email: body.Email, Name: name, Image: body.Image}, nil
}

// verifySignature só aceita o algoritmo que combina com o tipo da chave. É o que fecha a
// porta do clássico "alg: none" e da troca de algoritmo no cabeçalho.
func verifySignature(alg string, key crypto.PublicKey, signed string, signature []byte) bool {
	switch pub := key.(type) {
	case ed25519.PublicKey:
		return alg == "EdDSA" && ed25519.Verify(pub, []byte(signed), signature)
	case *rsa.PublicKey:
		digest := sha256.Sum256([]byte(signed))
		return alg == "RS256" && rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest[:], signature) == nil
	default:
		return false
	}
}

func decodeSegment(segment string, target any) error {
	raw, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, target)
}
