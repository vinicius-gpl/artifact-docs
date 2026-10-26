package auth

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

var testNow = time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

type fixedKeys map[string]crypto.PublicKey

func (keys fixedKeys) Key(_ context.Context, kid string) (crypto.PublicKey, error) {
	key, ok := keys[kid]
	if !ok {
		return nil, ErrUnknownKey
	}

	return key, nil
}

func segment(test *testing.T, value any) string {
	test.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		test.Fatalf("marshal: %v", err)
	}

	return base64.RawURLEncoding.EncodeToString(raw)
}

func signEd25519(test *testing.T, key ed25519.PrivateKey, head, body any) string {
	test.Helper()

	signed := segment(test, head) + "." + segment(test, body)

	return signed + "." + base64.RawURLEncoding.EncodeToString(ed25519.Sign(key, []byte(signed)))
}

func validClaims() map[string]any {
	return map[string]any{
		"sub":   "user-1",
		"email": "dono@exemplo.com",
		"name":  "Dono",
		"image": nil,
		"exp":   testNow.Add(10 * time.Minute).Unix(),
	}
}

func newTestVerifier(keys KeyFinder) *Verifier {
	verifier := NewVerifier(keys)
	verifier.now = func() time.Time { return testNow }

	return verifier
}

func newEd25519(test *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	test.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		test.Fatalf("generate key: %v", err)
	}

	return pub, priv
}

// feliz
func TestVerifyAcceptsAnEd25519TokenFromNeon(test *testing.T) {
	pub, priv := newEd25519(test)
	token := signEd25519(test, priv, map[string]string{"alg": "EdDSA", "kid": "k1"}, validClaims())

	user, err := newTestVerifier(fixedKeys{"k1": pub}).Verify(context.Background(), token)
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	if user.ID != "user-1" || user.Email != "dono@exemplo.com" || user.Name != "Dono" {
		test.Fatalf("user = %+v", user)
	}
}

// feliz
func TestVerifyAcceptsAnRS256Token(test *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		test.Fatalf("generate key: %v", err)
	}
	signed := segment(test, map[string]string{"alg": "RS256", "kid": "r1"}) + "." + segment(test, validClaims())
	digest := sha256.Sum256([]byte(signed))
	signature, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, digest[:])
	if err != nil {
		test.Fatalf("sign: %v", err)
	}
	token := signed + "." + base64.RawURLEncoding.EncodeToString(signature)

	if _, err := newTestVerifier(fixedKeys{"r1": &priv.PublicKey}).Verify(context.Background(), token); err != nil {
		test.Fatalf("unexpected error: %v", err)
	}
}

// feliz
func TestVerifyFallsBackToTheEmailWhenThereIsNoName(test *testing.T) {
	pub, priv := newEd25519(test)
	body := validClaims()
	delete(body, "name")
	token := signEd25519(test, priv, map[string]string{"alg": "EdDSA", "kid": "k1"}, body)

	user, err := newTestVerifier(fixedKeys{"k1": pub}).Verify(context.Background(), token)
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	if user.Name != "dono@exemplo.com" {
		test.Fatalf("name = %q, want the email", user.Name)
	}
}

// triste
func TestVerifyRejectsWhatDoesNotProveAnything(test *testing.T) {
	pub, priv := newEd25519(test)
	_, otherPriv := newEd25519(test)
	head := map[string]string{"alg": "EdDSA", "kid": "k1"}
	expired := validClaims()
	expired["exp"] = testNow.Add(-time.Second).Unix()
	noEmail := validClaims()
	delete(noEmail, "email")
	valid := signEd25519(test, priv, head, validClaims())

	cases := map[string]string{
		"malformed":       "not-a-jwt",
		"other signer":    signEd25519(test, otherPriv, head, validClaims()),
		"expired":         signEd25519(test, priv, head, expired),
		"without email":   signEd25519(test, priv, head, noEmail),
		"unknown kid":     signEd25519(test, priv, map[string]string{"alg": "EdDSA", "kid": "nope"}, validClaims()),
		"alg none":        segment(test, map[string]string{"alg": "none", "kid": "k1"}) + "." + segment(test, validClaims()) + ".",
		"tampered claims": valid[:len(valid)/2] + "x" + valid[len(valid)/2+1:],
	}

	verifier := newTestVerifier(fixedKeys{"k1": pub})
	for name, token := range cases {
		if _, err := verifier.Verify(context.Background(), token); !errors.Is(err, ErrInvalidToken) {
			test.Errorf("%s: err = %v, want ErrInvalidToken", name, err)
		}
	}
}

func jwksServer(test *testing.T, keys []jwk, hits *atomic.Int32) *httptest.Server {
	test.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_ = json.NewEncoder(writer).Encode(map[string]any{"keys": keys})
	}))
	test.Cleanup(srv.Close)

	return srv
}

// feliz
func TestKeySetFetchesOnceAndReadsBothKeyTypes(test *testing.T) {
	pub, _ := newEd25519(test)
	var hits atomic.Int32
	srv := jwksServer(test, []jwk{
		{Kid: "k1", Kty: "OKP", Crv: "Ed25519", PublicX: base64.RawURLEncoding.EncodeToString(pub)},
		{Kid: "r1", Kty: "RSA", Modulus: base64.RawURLEncoding.EncodeToString(big.NewInt(12345).Bytes()), Exponent: "AQAB"},
	}, &hits)
	keys := NewKeySet(srv.URL, srv.Client())

	for _, kid := range []string{"k1", "r1", "k1"} {
		if _, err := keys.Key(context.Background(), kid); err != nil {
			test.Fatalf("%s: unexpected error: %v", kid, err)
		}
	}

	if hits.Load() != 1 {
		test.Fatalf("jwks fetched %d times, want 1", hits.Load())
	}
}

// triste
func TestKeySetDoesNotHammerNeonWithUnknownKids(test *testing.T) {
	var hits atomic.Int32
	srv := jwksServer(test, nil, &hits)
	keys := NewKeySet(srv.URL, srv.Client())
	keys.now = func() time.Time { return testNow }

	for range 5 {
		if _, err := keys.Key(context.Background(), "forged"); !errors.Is(err, ErrUnknownKey) {
			test.Fatalf("err = %v, want ErrUnknownKey", err)
		}
	}
	if hits.Load() != 1 {
		test.Fatalf("jwks fetched %d times inside the interval, want 1", hits.Load())
	}

	keys.now = func() time.Time { return testNow.Add(minRefreshInterval) }
	_, _ = keys.Key(context.Background(), "forged")
	if hits.Load() != 2 {
		test.Fatalf("jwks fetched %d times after the interval, want 2", hits.Load())
	}
}

// triste
func TestKeySetReportsANeonOutage(test *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusBadGateway)
	}))
	test.Cleanup(srv.Close)

	_, err := NewKeySet(srv.URL, srv.Client()).Key(context.Background(), "k1")
	if err == nil || errors.Is(err, ErrUnknownKey) {
		test.Fatalf("err = %v, want a fetch error", err)
	}
}
