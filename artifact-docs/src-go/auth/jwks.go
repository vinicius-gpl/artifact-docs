// Package auth confere o token que o Neon Auth assina — a prova de quem está chamando a API.
//
// O Go nunca vê a senha: ela vai do navegador direto para a Neon, e de volta vem um JWT. Aqui
// só se confere a assinatura (com as chaves públicas da Neon, o JWKS) e a validade. Tudo com a
// biblioteca padrão: são duas dúzias de linhas de criptografia, e uma dependência a menos é
// memória e superfície de ataque a menos na VPS de 1 GB.
package auth

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// minRefreshInterval é o menor intervalo entre duas buscas do JWKS. Um token forjado com um
// `kid` inventado força uma busca nova; sem este freio, cada requisição dessas viraria uma
// chamada à Neon.
const minRefreshInterval = 30 * time.Second

// ErrUnknownKey é o `kid` que não está no JWKS nem depois de buscá-lo de novo.
var ErrUnknownKey = errors.New("auth: unknown signing key")

// KeySet guarda em memória as chaves públicas do Neon Auth.
//
// Busca uma vez, na primeira requisição, e só busca de novo quando aparece um `kid`
// desconhecido — é assim que a Neon troca de chave sem derrubar ninguém.
type KeySet struct {
	url    string
	client *http.Client
	now    func() time.Time

	mu        sync.Mutex
	keys      map[string]crypto.PublicKey
	fetchedAt time.Time
}

// NewKeySet aponta para o JWKS em url (no painel da Neon: Auth → JWKS URL).
func NewKeySet(url string, client *http.Client) *KeySet {
	return &KeySet{url: url, client: client, now: time.Now}
}

// Key devolve a chave pública com este kid.
func (keySet *KeySet) Key(ctx context.Context, kid string) (crypto.PublicKey, error) {
	keySet.mu.Lock()
	defer keySet.mu.Unlock()

	if key, ok := keySet.keys[kid]; ok {
		return key, nil
	}
	if keySet.keys != nil && keySet.now().Sub(keySet.fetchedAt) < minRefreshInterval {
		return nil, ErrUnknownKey
	}

	keys, err := keySet.fetch(ctx)
	if err != nil {
		return nil, err
	}
	keySet.keys = keys
	keySet.fetchedAt = keySet.now()

	if key, ok := keys[kid]; ok {
		return key, nil
	}

	return nil, ErrUnknownKey
}

func (keySet *KeySet) fetch(ctx context.Context) (map[string]crypto.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, keySet.url, nil)
	if err != nil {
		return nil, fmt.Errorf("auth: build jwks request: %w", err)
	}

	res, err := keySet.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth: fetch jwks: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth: fetch jwks: status %d", res.StatusCode)
	}

	var body struct {
		Keys []jwk `json:"keys"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("auth: decode jwks: %w", err)
	}

	return parseKeys(body.Keys), nil
}

// jwk é uma chave do JWKS. Só os campos das duas famílias que a Neon pode usar: Ed25519
// (OKP, o padrão do Better Auth) e RSA.
type jwk struct {
	Kid      string `json:"kid"`
	Kty      string `json:"kty"`
	Crv      string `json:"crv"`
	PublicX  string `json:"x"`
	Modulus  string `json:"n"`
	Exponent string `json:"e"`
}

// parseKeys ignora a chave que não sabe ler, em vez de recusar o JWKS inteiro: uma chave
// nova de outro tipo na Neon não pode derrubar o login de quem usa as que já funcionam.
func parseKeys(raw []jwk) map[string]crypto.PublicKey {
	keys := make(map[string]crypto.PublicKey, len(raw))
	for _, item := range raw {
		key, err := item.publicKey()
		if err != nil {
			continue
		}
		keys[item.Kid] = key
	}

	return keys
}

func (entry jwk) publicKey() (crypto.PublicKey, error) {
	if entry.Kty == "OKP" && entry.Crv == "Ed25519" {
		return ed25519Key(entry.PublicX)
	}
	if entry.Kty == "RSA" {
		return rsaKey(entry.Modulus, entry.Exponent)
	}

	return nil, fmt.Errorf("auth: unsupported key type %q", entry.Kty)
}

func ed25519Key(encodedX string) (crypto.PublicKey, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encodedX)
	if err != nil || len(raw) != ed25519.PublicKeySize {
		return nil, errors.New("auth: invalid Ed25519 key")
	}

	return ed25519.PublicKey(raw), nil
}

func rsaKey(encodedModulus, encodedExponent string) (crypto.PublicKey, error) {
	modulus, err := base64.RawURLEncoding.DecodeString(encodedModulus)
	if err != nil {
		return nil, errors.New("auth: invalid RSA modulus")
	}
	exponent, err := base64.RawURLEncoding.DecodeString(encodedExponent)
	if err != nil || len(exponent) == 0 || len(exponent) > 4 {
		return nil, errors.New("auth: invalid RSA exponent")
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulus),
		E: int(new(big.Int).SetBytes(exponent).Int64()),
	}, nil
}
