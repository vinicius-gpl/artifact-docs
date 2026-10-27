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
func TestLoadReadsTheAuthAndDatabaseSettings(test *testing.T) {
	cfg, err := Load(envOf(map[string]string{
		"NEON_AUTH_URL":       " https://auth.example.com ",
		"NEON_AUTH_JWKS_URL":  "https://auth.example.com/.well-known/jwks.json",
		"AUTH_ALLOWED_EMAILS": "Dono@Exemplo.com, outra@exemplo.com,,",
		"DATABASE_URL":        " postgres://neon.example/hub ",
	}))
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	if cfg.DatabaseURL != "postgres://neon.example/hub" {
		test.Fatalf("database url = %q", cfg.DatabaseURL)
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

// feliz
func TestLoadReadsTheStorageAndTheLinkSecret(test *testing.T) {
	cfg, err := Load(envOf(map[string]string{
		"R2_ENDPOINT":          "https://account.r2.cloudflarestorage.com",
		"R2_ACCESS_KEY_ID":     "key",
		"R2_SECRET_ACCESS_KEY": "secret",
		"R2_BUCKET":            "artifact-docs",
		"ARTIFACT_LINK_SECRET": strings.Repeat("s", 32),
	}))
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Storage.Configured() || cfg.Storage.Bucket != "artifact-docs" {
		test.Fatalf("storage = %+v", cfg.Storage)
	}
	if len(cfg.LinkSecret) != 32 {
		test.Fatalf("link secret = %q", cfg.LinkSecret)
	}
}

// triste
func TestLoadRefusesAHalfConfiguredStorageOrAShortSecret(test *testing.T) {
	halfStorage := map[string]string{"R2_ENDPOINT": "https://account.r2.cloudflarestorage.com", "R2_BUCKET": "artifact-docs"}
	if _, err := Load(envOf(halfStorage)); err == nil {
		test.Error("half the R2 settings: expected an error")
	}

	if _, err := Load(envOf(map[string]string{"ARTIFACT_LINK_SECRET": "short"})); err == nil {
		test.Error("short link secret: expected an error")
	}
}

// feliz
func TestLoadReadsTheEnvironmentAndDefaultsToDevelopment(test *testing.T) {
	standard, err := Load(envOf(nil))
	if err != nil || standard.Env != EnvDevelopment || standard.IsTest() {
		test.Fatalf("default: env = %q, err = %v", standard.Env, err)
	}

	e2e, err := Load(envOf(map[string]string{"APP_ENV": "test"}))
	if err != nil || !e2e.IsTest() {
		test.Fatalf("test: env = %q, err = %v", e2e.Env, err)
	}
}

// triste
func TestLoadRefusesAnUnknownEnvOrTestModeWiredToNeonAuth(test *testing.T) {
	if _, err := Load(envOf(map[string]string{"APP_ENV": "staging"})); err == nil {
		test.Error("unknown APP_ENV: expected an error")
	}

	wired := map[string]string{"APP_ENV": "test", "NEON_AUTH_JWKS_URL": "https://auth.example.com/jwks"}
	if _, err := Load(envOf(wired)); err == nil {
		test.Error("test mode with Neon Auth configured: expected an error")
	}
}
