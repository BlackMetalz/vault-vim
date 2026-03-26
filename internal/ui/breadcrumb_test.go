package ui

import "testing"

func TestRenderBreadcrumb_Empty(t *testing.T) {
	got := renderBreadcrumb([]string{})
	if got != "" {
		t.Errorf("empty parts should return empty, got %q", got)
	}
}

func TestRenderBreadcrumb_Single(t *testing.T) {
	got := renderBreadcrumb([]string{"Mounts"})
	if got == "" {
		t.Error("single part should return non-empty")
	}
}

func TestRenderBreadcrumb_Multiple(t *testing.T) {
	got := renderBreadcrumb([]string{"Mounts", "secret/", "myapp", "prod"})
	if got == "" {
		t.Error("multiple parts should return non-empty")
	}
}
