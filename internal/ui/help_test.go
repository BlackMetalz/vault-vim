package ui

import "testing"

func TestRenderHelp_NonEmpty(t *testing.T) {
	bindings := []helpBinding{
		{"enter", "select"},
		{"esc", "back"},
		{"q", "quit"},
	}
	got := renderHelp(bindings, 80)
	if got == "" {
		t.Error("renderHelp should return non-empty output")
	}
}

func TestRenderHelp_SmallWidth(t *testing.T) {
	bindings := []helpBinding{
		{"enter", "select"},
		{"esc", "back"},
		{"q", "quit"},
	}
	got := renderHelp(bindings, 20)
	if got == "" {
		t.Error("renderHelp should return non-empty output even at small width")
	}
}

func TestRenderHelpModal_Mounts(t *testing.T) {
	got := renderHelpModal(viewMounts, 80)
	if got == "" {
		t.Error("help modal for mounts should be non-empty")
	}
}

func TestRenderHelpModal_Browser(t *testing.T) {
	got := renderHelpModal(viewBrowser, 80)
	if got == "" {
		t.Error("help modal for browser should be non-empty")
	}
}

func TestRenderHelpModal_Secret(t *testing.T) {
	got := renderHelpModal(viewSecret, 80)
	if got == "" {
		t.Error("help modal for secret should be non-empty")
	}
}

func TestRenderHelpSection_NonEmpty(t *testing.T) {
	got := renderHelpSection("Test", []helpBinding{
		{"a", "action"},
	}, 60)
	if got == "" {
		t.Error("renderHelpSection should be non-empty")
	}
}
