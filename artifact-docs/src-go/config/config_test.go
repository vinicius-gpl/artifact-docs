package config

import "testing"

func envOf(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

// feliz
func TestLoadUsesTheDefaultPortWhenNoneIsSet(t *testing.T) {
	cfg, err := Load(envOf(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != DefaultPort {
		t.Fatalf("port = %d, want %d", cfg.Port, DefaultPort)
	}
	if cfg.Addr() != ":3010" {
		t.Fatalf("addr = %q, want %q", cfg.Addr(), ":3010")
	}
}

// feliz
func TestLoadReadsThePortFromTheEnvironment(t *testing.T) {
	cfg, err := Load(envOf(map[string]string{"API_PORT": "8080"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Fatalf("port = %d, want 8080", cfg.Port)
	}
}

// triste
func TestLoadRejectsAnInvalidPort(t *testing.T) {
	for _, raw := range []string{"abc", "0", "70000", "-1"} {
		if _, err := Load(envOf(map[string]string{"API_PORT": raw})); err == nil {
			t.Errorf("API_PORT=%q: expected an error", raw)
		}
	}
}
