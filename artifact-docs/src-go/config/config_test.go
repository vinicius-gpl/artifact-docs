package config

import (
	"strings"
	"testing"
)

func envOf(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

// feliz
func TestLoadUsesTheDefaultPortWhenNoneIsSet(test *testing.T) {
	cfg, err := Load(envOf(nil))
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != DefaultPort {
		test.Fatalf("port = %d, want %d", cfg.Port, DefaultPort)
	}
	if cfg.Addr() != ":3010" {
		test.Fatalf("addr = %q, want %q", cfg.Addr(), ":3010")
	}
}

// feliz
func TestLoadReadsThePortFromTheEnvironment(test *testing.T) {
	cfg, err := Load(envOf(map[string]string{"API_PORT": "8080"}))
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		test.Fatalf("port = %d, want 8080", cfg.Port)
	}
}

// triste
func TestLoadRejectsAnInvalidPort(test *testing.T) {
	for _, raw := range []string{"abc", "0", "70000", "-1"} {
		if _, err := Load(envOf(map[string]string{"API_PORT": raw})); err == nil {
			test.Errorf("API_PORT=%q: expected an error", raw)
		}
	}
}

// feliz
func TestLoadReadsTheAuthSettings(test *testing.T) {
	cfg, err := Load(envOf(map[string]string{
		"NEON_AUTH_URL":       " https://auth.example.com ",
		"NEON_AUTH_JWKS_URL":  "https://auth.example.com/.well-known/jwks.json",
		"AUTH_ALLOWED_EMAILS": "Dono@Exemplo.com, outra@exemplo.com,,",
	}))
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	if cfg.NeonAuthURL != "https://auth.example.com" {
		test.Fatalf("neon auth url = %q", cfg.NeonAuthURL)
	}
	if cfg.NeonAuthJWKSURL != "https://auth.example.com/.well-known/jwks.json" {
		test.Fatalf("jwks url = %q", cfg.NeonAuthJWKSURL)
	}
	want := []string{"dono@exemplo.com", "outra@exemplo.com"}
	if strings.Join(cfg.AllowedEmails, "|") != strings.Join(want, "|") {
		test.Fatalf("allowed emails = %v, want %v", cfg.AllowedEmails, want)
	}
}

// triste
func TestLoadWithoutAllowedEmailsLetsNobodyIn(test *testing.T) {
	cfg, err := Load(envOf(map[string]string{"AUTH_ALLOWED_EMAILS": " , "}))
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.AllowedEmails) != 0 {
		test.Fatalf("allowed emails = %v, want none", cfg.AllowedEmails)
	}
}
