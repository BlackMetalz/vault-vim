package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Helper to create a KeyMsg for a rune character like 'a', 'C', etc.
func keyRunes(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// Helper to create a KeyMsg for a special key type.
func keyType(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

// --- startXxx method tests ---

func TestEditorNewOverlay(t *testing.T) {
	e := newEditorOverlay()
	if e.active {
		t.Error("new editor should not be active")
	}
}

func TestEditorStartEdit(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startEdit("db_host", "localhost:5432")

	if !e.active {
		t.Error("editor should be active")
	}
	if e.state != editorEditValue {
		t.Errorf("state: got %d, want %d", e.state, editorEditValue)
	}
	if e.key != "db_host" {
		t.Errorf("key: got %q, want %q", e.key, "db_host")
	}
	if e.textarea.Value() != "localhost:5432" {
		t.Errorf("value: got %q, want %q", e.textarea.Value(), "localhost:5432")
	}
}

func TestEditorStartAdd(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startAdd()

	if !e.active {
		t.Error("editor should be active")
	}
	if e.state != editorAddKeyName {
		t.Errorf("state: got %d, want %d", e.state, editorAddKeyName)
	}
	if e.key != "" {
		t.Errorf("key should be empty, got %q", e.key)
	}
	if e.input.Value() != "" {
		t.Errorf("input should be empty, got %q", e.input.Value())
	}
}

func TestEditorStartDelete(t *testing.T) {
	e := newEditorOverlay()
	e = e.startDelete("secret_key")

	if !e.active {
		t.Error("editor should be active")
	}
	if e.state != editorConfirmDelete {
		t.Errorf("state: got %d, want %d", e.state, editorConfirmDelete)
	}
	if e.key != "secret_key" {
		t.Errorf("key: got %q, want %q", e.key, "secret_key")
	}
	if e.input.Value() != "" {
		t.Errorf("input should be empty, got %q", e.input.Value())
	}
}

func TestEditorStartDeletePath(t *testing.T) {
	e := newEditorOverlay()

	t.Run("file", func(t *testing.T) {
		e2 := e.startDeletePath("secret/", "myapp/config", false)
		if !e2.active {
			t.Error("editor should be active")
		}
		if e2.state != editorConfirmDeletePath {
			t.Errorf("state: got %d, want %d", e2.state, editorConfirmDeletePath)
		}
		if e2.newSecMount != "secret/" {
			t.Errorf("mount: got %q, want %q", e2.newSecMount, "secret/")
		}
		if e2.newSecPath != "myapp/config" {
			t.Errorf("path: got %q, want %q", e2.newSecPath, "myapp/config")
		}
		if e2.key != "myapp/config" {
			t.Errorf("key: got %q, want %q", e2.key, "myapp/config")
		}
	})

	t.Run("directory", func(t *testing.T) {
		e2 := e.startDeletePath("secret/", "myapp/", true)
		if e2.key != "myapp/ (folder and all secrets inside)" {
			t.Errorf("key: got %q, want %q", e2.key, "myapp/ (folder and all secrets inside)")
		}
	})
}

func TestEditorStartRename(t *testing.T) {
	e := newEditorOverlay()
	e = e.startRename("old_key")

	if !e.active {
		t.Error("editor should be active")
	}
	if e.state != editorRenameKey {
		t.Errorf("state: got %d, want %d", e.state, editorRenameKey)
	}
	if e.oldKey != "old_key" {
		t.Errorf("oldKey: got %q, want %q", e.oldKey, "old_key")
	}
	if e.key != "old_key" {
		t.Errorf("key: got %q, want %q", e.key, "old_key")
	}
	if e.input.Value() != "old_key" {
		t.Errorf("input value: got %q, want %q", e.input.Value(), "old_key")
	}
}

func TestEditorStartNewSecret(t *testing.T) {
	e := newEditorOverlay()
	e = e.startNewSecret("secret/", "apps/")

	if !e.active {
		t.Error("editor should be active")
	}
	if e.state != editorNewSecretName {
		t.Errorf("state: got %d, want %d", e.state, editorNewSecretName)
	}
	if e.newSecMount != "secret/" {
		t.Errorf("mount: got %q, want %q", e.newSecMount, "secret/")
	}
	if e.newSecPath != "apps/" {
		t.Errorf("path: got %q, want %q", e.newSecPath, "apps/")
	}
	if e.input.Value() != "" {
		t.Errorf("input should be empty, got %q", e.input.Value())
	}
}

func TestEditorStartNewFolder(t *testing.T) {
	e := newEditorOverlay()
	e = e.startNewFolder("secret/", "apps/")

	if !e.active {
		t.Error("editor should be active")
	}
	if e.state != editorNewFolderName {
		t.Errorf("state: got %d, want %d", e.state, editorNewFolderName)
	}
	if e.newSecMount != "secret/" {
		t.Errorf("mount: got %q, want %q", e.newSecMount, "secret/")
	}
	if e.newSecPath != "apps/" {
		t.Errorf("path: got %q, want %q", e.newSecPath, "apps/")
	}
}

// --- updateConfirmDelete tests ---

func TestUpdateConfirmDelete_EscCancels(t *testing.T) {
	e := newEditorOverlay()
	e = e.startDelete("mykey")

	e, cmd := e.updateConfirmDelete(keyType(tea.KeyEsc))
	if e.active {
		t.Error("editor should be inactive after esc")
	}
	if cmd == nil {
		t.Fatal("cmd should not be nil")
	}
	msg := cmd()
	done, ok := msg.(editorDoneMsg)
	if !ok {
		t.Fatalf("expected editorDoneMsg, got %T", msg)
	}
	if done.action != "cancel" {
		t.Errorf("action: got %q, want %q", done.action, "cancel")
	}
}

func TestUpdateConfirmDelete_WrongTextDoesNothing(t *testing.T) {
	e := newEditorOverlay()
	e = e.startDelete("mykey")
	e.input.SetValue("wrong")

	e, cmd := e.updateConfirmDelete(keyType(tea.KeyEnter))
	if !e.active {
		t.Error("editor should remain active when wrong text entered")
	}
	if cmd != nil {
		t.Error("cmd should be nil when wrong confirmation text")
	}
}

func TestUpdateConfirmDelete_CorrectTextDeletes(t *testing.T) {
	e := newEditorOverlay()
	e = e.startDelete("mykey")
	e.input.SetValue("Confirm")

	e, cmd := e.updateConfirmDelete(keyType(tea.KeyEnter))
	if e.active {
		t.Error("editor should be inactive after confirmed delete")
	}
	if cmd == nil {
		t.Fatal("cmd should not be nil")
	}
	msg := cmd()
	done, ok := msg.(editorDoneMsg)
	if !ok {
		t.Fatalf("expected editorDoneMsg, got %T", msg)
	}
	if done.action != "delete" {
		t.Errorf("action: got %q, want %q", done.action, "delete")
	}
	if done.key != "mykey" {
		t.Errorf("key: got %q, want %q", done.key, "mykey")
	}
}

func TestUpdateConfirmDeletePath_CorrectTextDeletesPath(t *testing.T) {
	e := newEditorOverlay()
	e = e.startDeletePath("secret/", "myapp/config", false)
	e.input.SetValue("Confirm")

	e, cmd := e.updateConfirmDelete(keyType(tea.KeyEnter))
	if e.active {
		t.Error("editor should be inactive")
	}
	if cmd == nil {
		t.Fatal("cmd should not be nil")
	}
	msg := cmd()
	done, ok := msg.(editorDoneMsg)
	if !ok {
		t.Fatalf("expected editorDoneMsg, got %T", msg)
	}
	if done.action != "deletepath" {
		t.Errorf("action: got %q, want %q", done.action, "deletepath")
	}
	if done.key != "secret/" {
		t.Errorf("key (mount): got %q, want %q", done.key, "secret/")
	}
	if done.value != "myapp/config" {
		t.Errorf("value (path): got %q, want %q", done.value, "myapp/config")
	}
}

// --- updateKeyNameInput tests ---

func TestUpdateKeyNameInput_EnterWithEmptyDoesNothing(t *testing.T) {
	e := newEditorOverlay()
	e = e.startAdd()

	e, cmd := e.updateKeyNameInput(keyType(tea.KeyEnter))
	if e.state != editorAddKeyName {
		t.Errorf("state should remain editorAddKeyName, got %d", e.state)
	}
	if cmd != nil {
		t.Error("cmd should be nil for empty name")
	}
}

func TestUpdateKeyNameInput_EnterWithNameTransitions(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startAdd()
	e.input.SetValue("new_key")

	e, cmd := e.updateKeyNameInput(keyType(tea.KeyEnter))
	if e.state != editorAddKeyValue {
		t.Errorf("state: got %d, want %d (editorAddKeyValue)", e.state, editorAddKeyValue)
	}
	if e.key != "new_key" {
		t.Errorf("key: got %q, want %q", e.key, "new_key")
	}
	if cmd != nil {
		t.Error("cmd should be nil (no message needed for state transition)")
	}
}

func TestUpdateKeyNameInput_EscCancels(t *testing.T) {
	e := newEditorOverlay()
	e = e.startAdd()
	e.input.SetValue("partial")

	e, cmd := e.updateKeyNameInput(keyType(tea.KeyEsc))
	if e.active {
		t.Error("editor should be inactive after esc")
	}
	if cmd == nil {
		t.Fatal("cmd should not be nil")
	}
	msg := cmd()
	done, ok := msg.(editorDoneMsg)
	if !ok {
		t.Fatalf("expected editorDoneMsg, got %T", msg)
	}
	if done.action != "cancel" {
		t.Errorf("action: got %q, want %q", done.action, "cancel")
	}
}

// --- updateRenameInput tests ---

func TestUpdateRenameInput_EnterWithSameNameCancels(t *testing.T) {
	e := newEditorOverlay()
	e = e.startRename("old_key")
	// input already has "old_key" from startRename

	e, cmd := e.updateRenameInput(keyType(tea.KeyEnter))
	if e.active {
		t.Error("editor should be inactive when name unchanged")
	}
	if cmd == nil {
		t.Fatal("cmd should not be nil")
	}
	msg := cmd()
	done, ok := msg.(editorDoneMsg)
	if !ok {
		t.Fatalf("expected editorDoneMsg, got %T", msg)
	}
	if done.action != "cancel" {
		t.Errorf("action: got %q, want %q", done.action, "cancel")
	}
}

func TestUpdateRenameInput_EnterWithEmptyNameCancels(t *testing.T) {
	e := newEditorOverlay()
	e = e.startRename("old_key")
	e.input.SetValue("")

	e, cmd := e.updateRenameInput(keyType(tea.KeyEnter))
	if e.active {
		t.Error("editor should be inactive when name is empty")
	}
	if cmd == nil {
		t.Fatal("cmd should not be nil")
	}
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "cancel" {
		t.Errorf("action: got %q, want %q", done.action, "cancel")
	}
}

func TestUpdateRenameInput_EnterWithNewNameRenames(t *testing.T) {
	e := newEditorOverlay()
	e = e.startRename("old_key")
	e.input.SetValue("new_key")

	e, cmd := e.updateRenameInput(keyType(tea.KeyEnter))
	if e.active {
		t.Error("editor should be inactive after rename")
	}
	if cmd == nil {
		t.Fatal("cmd should not be nil")
	}
	msg := cmd()
	done, ok := msg.(editorDoneMsg)
	if !ok {
		t.Fatalf("expected editorDoneMsg, got %T", msg)
	}
	if done.action != "rename" {
		t.Errorf("action: got %q, want %q", done.action, "rename")
	}
	if done.key != "new_key" {
		t.Errorf("key (new name): got %q, want %q", done.key, "new_key")
	}
	if done.value != "old_key" {
		t.Errorf("value (old name): got %q, want %q", done.value, "old_key")
	}
}

func TestUpdateRenameInput_EscCancels(t *testing.T) {
	e := newEditorOverlay()
	e = e.startRename("old_key")

	e, cmd := e.updateRenameInput(keyType(tea.KeyEsc))
	if e.active {
		t.Error("editor should be inactive after esc")
	}
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "cancel" {
		t.Errorf("action: got %q, want %q", done.action, "cancel")
	}
}

// --- updateNewSecretName tests ---

func TestUpdateNewSecretName_EnterWithEmptyDoesNothing(t *testing.T) {
	e := newEditorOverlay()
	e = e.startNewSecret("secret/", "apps/")

	e, cmd := e.updateNewSecretName(keyType(tea.KeyEnter))
	if !e.active {
		t.Error("editor should remain active with empty name")
	}
	if cmd != nil {
		t.Error("cmd should be nil for empty name")
	}
}

func TestUpdateNewSecretName_EnterWithNameReturnsAction(t *testing.T) {
	e := newEditorOverlay()
	e = e.startNewSecret("secret/", "apps/")
	e.input.SetValue("my-app")

	e, cmd := e.updateNewSecretName(keyType(tea.KeyEnter))
	if e.active {
		t.Error("editor should be inactive after creating")
	}
	if cmd == nil {
		t.Fatal("cmd should not be nil")
	}
	msg := cmd()
	done, ok := msg.(editorDoneMsg)
	if !ok {
		t.Fatalf("expected editorDoneMsg, got %T", msg)
	}
	if done.action != "newsecret" {
		t.Errorf("action: got %q, want %q", done.action, "newsecret")
	}
	if done.key != "secret/" {
		t.Errorf("key (mount): got %q, want %q", done.key, "secret/")
	}
	if done.value != "apps/my-app" {
		t.Errorf("value (path): got %q, want %q", done.value, "apps/my-app")
	}
}

func TestUpdateNewSecretName_EscCancels(t *testing.T) {
	e := newEditorOverlay()
	e = e.startNewSecret("secret/", "apps/")

	e, cmd := e.updateNewSecretName(keyType(tea.KeyEsc))
	if e.active {
		t.Error("editor should be inactive after esc")
	}
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "cancel" {
		t.Errorf("action: got %q, want %q", done.action, "cancel")
	}
}

// --- updateNewFolderName tests ---

func TestUpdateNewFolderName_EnterWithEmptyDoesNothing(t *testing.T) {
	e := newEditorOverlay()
	e = e.startNewFolder("secret/", "apps/")

	e, cmd := e.updateNewFolderName(keyType(tea.KeyEnter))
	if !e.active {
		t.Error("editor should remain active with empty name")
	}
	if cmd != nil {
		t.Error("cmd should be nil for empty name")
	}
}

func TestUpdateNewFolderName_EnterWithNameReturnsAction(t *testing.T) {
	e := newEditorOverlay()
	e = e.startNewFolder("secret/", "apps/")
	e.input.SetValue("subfolder")

	e, cmd := e.updateNewFolderName(keyType(tea.KeyEnter))
	if e.active {
		t.Error("editor should be inactive after creating")
	}
	if cmd == nil {
		t.Fatal("cmd should not be nil")
	}
	msg := cmd()
	done, ok := msg.(editorDoneMsg)
	if !ok {
		t.Fatalf("expected editorDoneMsg, got %T", msg)
	}
	if done.action != "newfolder" {
		t.Errorf("action: got %q, want %q", done.action, "newfolder")
	}
	if done.key != "secret/" {
		t.Errorf("key (mount): got %q, want %q", done.key, "secret/")
	}
	if done.value != "apps/subfolder" {
		t.Errorf("value (path): got %q, want %q", done.value, "apps/subfolder")
	}
}

func TestUpdateNewFolderName_EscCancels(t *testing.T) {
	e := newEditorOverlay()
	e = e.startNewFolder("secret/", "apps/")

	e, cmd := e.updateNewFolderName(keyType(tea.KeyEsc))
	if e.active {
		t.Error("editor should be inactive after esc")
	}
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "cancel" {
		t.Errorf("action: got %q, want %q", done.action, "cancel")
	}
}

// --- updateTextarea tests ---

func TestUpdateTextarea_EscCancels(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startEdit("key", "value")

	e, cmd := e.updateTextarea(keyType(tea.KeyEsc))
	if e.active {
		t.Error("editor should be inactive after esc")
	}
	if cmd == nil {
		t.Fatal("cmd should not be nil")
	}
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "cancel" {
		t.Errorf("action: got %q, want %q", done.action, "cancel")
	}
}

func TestUpdateTextarea_CtrlSSaves(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startEdit("db_host", "localhost")

	e, cmd := e.updateTextarea(keyType(tea.KeyCtrlS))
	if e.active {
		t.Error("editor should be inactive after save")
	}
	if cmd == nil {
		t.Fatal("cmd should not be nil")
	}
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "edit" {
		t.Errorf("action: got %q, want %q", done.action, "edit")
	}
	if done.key != "db_host" {
		t.Errorf("key: got %q, want %q", done.key, "db_host")
	}
}

func TestUpdateTextarea_F2Saves(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startEdit("key", "val")

	e, cmd := e.updateTextarea(keyType(tea.KeyF2))
	if e.active {
		t.Error("editor should be inactive after F2 save")
	}
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "edit" {
		t.Errorf("action: got %q, want %q", done.action, "edit")
	}
}

func TestUpdateTextarea_CtrlSSavesAddMode(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startAdd()
	e.input.SetValue("new_key")
	// Simulate transitioning to addKeyValue state
	e, _ = e.updateKeyNameInput(keyType(tea.KeyEnter))
	e.textarea.SetValue("new_value")

	e, cmd := e.updateTextarea(keyType(tea.KeyCtrlS))
	if e.active {
		t.Error("editor should be inactive after save")
	}
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "add" {
		t.Errorf("action: got %q, want %q", done.action, "add")
	}
	if done.key != "new_key" {
		t.Errorf("key: got %q, want %q", done.key, "new_key")
	}
}

func TestUpdateTextarea_CtrlDDeletesLine(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startEdit("key", "line1\nline2\nline3")

	// textarea.Line() returns the cursor's current line.
	// After SetValue, cursor may be at any line depending on textarea internals.
	// We just verify that one line gets removed and the result has 2 lines.
	e, cmd := e.updateTextarea(keyType(tea.KeyCtrlD))
	if cmd != nil {
		t.Error("ctrl+d should not produce a cmd")
	}
	got := e.textarea.Value()
	lines := splitLines(got)
	if len(lines) != 2 {
		t.Errorf("after ctrl+d: expected 2 lines, got %d (value: %q)", len(lines), got)
	}
}

func TestUpdateTextarea_CtrlKSwallowed(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startEdit("key", "some value")

	original := e.textarea.Value()
	e, cmd := e.updateTextarea(keyType(tea.KeyCtrlK))
	if cmd != nil {
		t.Error("ctrl+k should not produce a cmd")
	}
	if e.textarea.Value() != original {
		t.Error("ctrl+k should not modify the value")
	}
}

func TestUpdateTextarea_CtrlUSwallowed(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startEdit("key", "some value")

	original := e.textarea.Value()
	e, cmd := e.updateTextarea(keyType(tea.KeyCtrlU))
	if cmd != nil {
		t.Error("ctrl+u should not produce a cmd")
	}
	if e.textarea.Value() != original {
		t.Error("ctrl+u should not modify the value")
	}
}

// --- Textarea SetValue/Value tests ---

func TestEditorSetValue_SingleLine(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startEdit("key", "")

	e.textarea.SetValue("hello world")
	if got := e.textarea.Value(); got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestEditorSetValue_Multiline(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startEdit("key", "")

	json := "{\n  \"username\": \"kienlt\"\n}"
	e.textarea.SetValue(json)
	if got := e.textarea.Value(); got != json {
		t.Errorf("SetValue multiline:\ngot:  %q\nwant: %q", got, json)
	}
}

func TestEditorSetValue_Empty(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startEdit("key", "initial")

	e.textarea.SetValue("")
	if got := e.textarea.Value(); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

// --- Update routing tests ---

func TestEditorUpdate_RoutesToConfirmDelete(t *testing.T) {
	e := newEditorOverlay()
	e = e.startDelete("mykey")
	e.input.SetValue("Confirm")

	e, cmd := e.Update(keyType(tea.KeyEnter))
	if e.active {
		t.Error("editor should be inactive after confirmed delete via Update")
	}
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "delete" {
		t.Errorf("action: got %q, want %q", done.action, "delete")
	}
}

func TestEditorUpdate_RoutesToAddKeyName(t *testing.T) {
	e := newEditorOverlay()
	e = e.startAdd()

	// Enter with empty should do nothing
	e, cmd := e.Update(keyType(tea.KeyEnter))
	if e.state != editorAddKeyName {
		t.Errorf("state should remain editorAddKeyName, got %d", e.state)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestEditorUpdate_RoutesToRename(t *testing.T) {
	e := newEditorOverlay()
	e = e.startRename("old")
	e.input.SetValue("new")

	e, cmd := e.Update(keyType(tea.KeyEnter))
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "rename" {
		t.Errorf("action: got %q, want %q", done.action, "rename")
	}
}

func TestEditorUpdate_RoutesToNewSecretName(t *testing.T) {
	e := newEditorOverlay()
	e = e.startNewSecret("secret/", "")
	e.input.SetValue("myapp")

	e, cmd := e.Update(keyType(tea.KeyEnter))
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "newsecret" {
		t.Errorf("action: got %q, want %q", done.action, "newsecret")
	}
}

func TestEditorUpdate_RoutesToNewFolderName(t *testing.T) {
	e := newEditorOverlay()
	e = e.startNewFolder("secret/", "")
	e.input.SetValue("myfolder")

	e, cmd := e.Update(keyType(tea.KeyEnter))
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "newfolder" {
		t.Errorf("action: got %q, want %q", done.action, "newfolder")
	}
}

func TestEditorUpdate_RoutesToTextarea(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startEdit("key", "value")

	e, cmd := e.Update(keyType(tea.KeyEsc))
	if e.active {
		t.Error("editor should be inactive after esc in textarea mode")
	}
	msg := cmd()
	done := msg.(editorDoneMsg)
	if done.action != "cancel" {
		t.Errorf("action: got %q, want %q", done.action, "cancel")
	}
}

// --- View tests ---

func TestEditorView_InactiveReturnsEmpty(t *testing.T) {
	e := newEditorOverlay()
	if v := e.View(); v != "" {
		t.Errorf("inactive editor should return empty view, got %q", v)
	}
}

func TestEditorView_ActiveReturnsContent(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e = e.startEdit("mykey", "myvalue")

	v := e.View()
	if v == "" {
		t.Error("active editor should return non-empty view")
	}
}

// --- splitLines / joinLines tests ---

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 1},
		{"hello", 1},
		{"a\nb\nc", 3},
	}
	for _, tt := range tests {
		got := splitLines(tt.input)
		if len(got) != tt.want {
			t.Errorf("splitLines(%q): got %d parts, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestJoinLines(t *testing.T) {
	got := joinLines([]string{"a", "b", "c"})
	if got != "a\nb\nc" {
		t.Errorf("joinLines: got %q, want %q", got, "a\nb\nc")
	}
}

// --- itoa tests ---

func TestItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{100, "100"},
	}
	for _, tt := range tests {
		got := itoa(tt.input)
		if got != tt.want {
			t.Errorf("itoa(%d): got %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- resizeTextarea tests ---

func TestResizeTextarea(t *testing.T) {
	e := newEditorOverlay()
	e.width = 80
	e.height = 40
	e.resizeTextarea()
	// Should not panic and should set reasonable dimensions
}

func TestResizeTextarea_SmallDimensions(t *testing.T) {
	e := newEditorOverlay()
	e.width = 10
	e.height = 10
	e.resizeTextarea()
	// Minimum width is 30, minimum height is 5 - should not panic
}
