package ui

import (
	"fmt"
	"strings"
	"testing"
)

func TestFlattenValue_SingleLine(t *testing.T) {
	got := flattenValue("hello world", 50)
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestFlattenValue_SingleLineTruncate(t *testing.T) {
	got := flattenValue("a very long single line value that exceeds the max", 20)
	if len(got) > 20 {
		t.Errorf("expected length <= 20, got %d: %q", len(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected truncation with ..., got %q", got)
	}
}

func TestFlattenValue_Multiline(t *testing.T) {
	input := "{\n  \"key\": \"value\"\n}"
	got := flattenValue(input, 50)
	if strings.Contains(got, "\n") {
		t.Errorf("multiline should be flattened, got %q", got)
	}
	if !strings.Contains(got, "\u21b5") {
		t.Errorf("should contain newline symbol, got %q", got)
	}
}

func TestFlattenValue_MultilineTruncate(t *testing.T) {
	input := "{\n  \"username\": \"kienlt\",\n  \"demo-key\": \"demo-value\"\n}"
	got := flattenValue(input, 30)
	if len(got) > 30 {
		t.Errorf("expected length <= 30, got %d: %q", len(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected truncation with ..., got %q", got)
	}
}

func TestFlattenValue_Empty(t *testing.T) {
	got := flattenValue("", 50)
	if got != "" {
		t.Errorf("empty input should return empty, got %q", got)
	}
}

func TestFlattenValue_ShortSingleLine(t *testing.T) {
	got := flattenValue("ok", 50)
	if got != "ok" {
		t.Errorf("got %q, want %q", got, "ok")
	}
}

func TestSecretLoadedMsg_404TreatedAsEmpty(t *testing.T) {
	m := newSecretModel(nil, "secret/", "nonexistent/path")
	m.width = 80
	m.height = 40

	// Simulate a 404 error from Vault
	err := fmt.Errorf("get secret secret/nonexistent/path: vault API error (HTTP 404): {\"errors\":[]}")
	msg := secretLoadedMsg{secret: nil, err: err}

	m, _ = m.Update(msg)

	if m.err != nil {
		t.Errorf("404 should not set error, got: %v", m.err)
	}
	if m.values == nil {
		t.Error("values should be initialized (empty map), not nil")
	}
	if len(m.keys) != 0 {
		t.Errorf("keys should be empty, got %d", len(m.keys))
	}
}

func TestSecretLoadedMsg_RealError(t *testing.T) {
	m := newSecretModel(nil, "secret/", "some/path")
	m.width = 80
	m.height = 40

	err := fmt.Errorf("connection refused")
	msg := secretLoadedMsg{secret: nil, err: err}

	m, _ = m.Update(msg)

	if m.err == nil {
		t.Error("non-404 error should be preserved")
	}
}

func TestSecretBreadcrumbs(t *testing.T) {
	m := newSecretModel(nil, "secret/", "myapp/prod")
	got := m.breadcrumbs()
	want := []string{"Mounts", "secret/", "myapp", "prod"}
	if len(got) != len(want) {
		t.Fatalf("breadcrumbs length: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("breadcrumbs[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}
