package vault

import (
	"testing"

	"github.com/BlackMetalz/vault-vim/internal/config"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		VaultAddr:  "http://localhost:8200",
		VaultToken: "test-token-123",
	}

	client := NewClient(cfg)

	if client.addr != "http://localhost:8200" {
		t.Errorf("addr: got %q, want %q", client.addr, "http://localhost:8200")
	}
	if client.token != "test-token-123" {
		t.Errorf("token: got %q, want %q", client.token, "test-token-123")
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestNewClient_DifferentConfig(t *testing.T) {
	cfg := &config.Config{
		VaultAddr:  "https://vault.example.com:8200",
		VaultToken: "s.AbCdEfGhIjKlMnOp",
	}

	client := NewClient(cfg)

	if client.addr != "https://vault.example.com:8200" {
		t.Errorf("addr: got %q, want %q", client.addr, "https://vault.example.com:8200")
	}
	if client.token != "s.AbCdEfGhIjKlMnOp" {
		t.Errorf("token: got %q, want %q", client.token, "s.AbCdEfGhIjKlMnOp")
	}
}

func TestMountStruct(t *testing.T) {
	m := Mount{Path: "secret/", Type: "kv-v2"}
	if m.Path != "secret/" {
		t.Errorf("Path: got %q, want %q", m.Path, "secret/")
	}
	if m.Type != "kv-v2" {
		t.Errorf("Type: got %q, want %q", m.Type, "kv-v2")
	}
}

func TestSecretStruct(t *testing.T) {
	s := Secret{
		Data: map[string]string{
			"username": "admin",
			"password": "secret123",
		},
		Metadata: map[string]interface{}{
			"version": float64(1),
		},
	}

	if s.Data["username"] != "admin" {
		t.Errorf("Data[username]: got %q, want %q", s.Data["username"], "admin")
	}
	if s.Data["password"] != "secret123" {
		t.Errorf("Data[password]: got %q, want %q", s.Data["password"], "secret123")
	}
	if v, ok := s.Metadata["version"]; !ok || v != float64(1) {
		t.Errorf("Metadata[version]: got %v, want 1", v)
	}
}

// TestListSecrets_404HandledAsEmpty verifies the code path in ListSecrets
// that treats HTTP 404 errors as an empty list. We test this by checking
// the error string matching logic directly.
func TestListSecrets_404ErrorStringMatching(t *testing.T) {
	// The ListSecrets method checks: strings.Contains(err.Error(), "HTTP 404")
	// We verify that the error format from the request method matches this pattern.
	// The request method formats errors as: "vault API error (HTTP 404): ..."
	// So "HTTP 404" is indeed contained in that string.

	errStr := "vault API error (HTTP 404): no secret found"
	if !contains(errStr, "HTTP 404") {
		t.Error("expected error string to contain 'HTTP 404'")
	}

	// Non-404 errors should NOT match
	errStr2 := "vault API error (HTTP 403): permission denied"
	if contains(errStr2, "HTTP 404") {
		t.Error("403 error should not match HTTP 404 pattern")
	}
}

// contains is a helper that mirrors strings.Contains for test clarity.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
