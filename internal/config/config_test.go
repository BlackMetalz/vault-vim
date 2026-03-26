package config

import (
	"os"
	"testing"
)

func TestLoad_WithEnvVars(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://test:8200")
	t.Setenv("VAULT_TOKEN", "test-token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.VaultAddr != "http://test:8200" {
		t.Errorf("VaultAddr: got %q, want %q", cfg.VaultAddr, "http://test:8200")
	}
	if cfg.VaultToken != "test-token" {
		t.Errorf("VaultToken: got %q, want %q", cfg.VaultToken, "test-token")
	}
}

func TestLoad_DefaultAddr(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "test-token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.VaultAddr != "http://127.0.0.1:8200" {
		t.Errorf("VaultAddr: got %q, want %q", cfg.VaultAddr, "http://127.0.0.1:8200")
	}
}

func TestLoad_NoToken(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://test:8200")
	t.Setenv("VAULT_TOKEN", "")

	// Ensure no vault-token file interferes
	home, _ := os.UserHomeDir()
	orig, err := os.ReadFile(home + "/.vault-token")
	if err == nil {
		os.Remove(home + "/.vault-token")
		defer os.WriteFile(home+"/.vault-token", orig, 0600)
	}

	_, err = Load()
	if err == nil {
		t.Error("expected error when no token available")
	}
}
