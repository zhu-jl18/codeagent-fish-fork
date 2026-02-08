package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func createPromptBaseForTest(t *testing.T) string {
	t.Helper()
	setRuntimeSettingsForTest(map[string]string{})
	t.Cleanup(resetRuntimeSettingsForTest)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	promptBase := filepath.Join(home, ".fish-agent-wrapper", "prompts")
	if err := os.MkdirAll(promptBase, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	return promptBase
}

func TestWrapTaskWithAgentPrompt(t *testing.T) {
	got := wrapTaskWithAgentPrompt("P", "do")
	want := "<agent-prompt>\nP\n</agent-prompt>\n\ndo"
	if got != want {
		t.Fatalf("wrapTaskWithAgentPrompt mismatch:\n got=%q\nwant=%q", got, want)
	}
}

func TestReadAgentPromptFile_EmptyPath(t *testing.T) {
	got, err := readAgentPromptFile("   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty result, got %q", got)
	}
}

func TestReadAgentPromptFile_AbsolutePathInsideClaudeDir(t *testing.T) {
	promptBase := createPromptBaseForTest(t)

	path := filepath.Join(promptBase, "prompt.md")
	if err := os.WriteFile(path, []byte("LINE1\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := readAgentPromptFile(path)
	if err != nil {
		t.Fatalf("readAgentPromptFile error: %v", err)
	}
	if got != "LINE1" {
		t.Fatalf("got %q, want %q", got, "LINE1")
	}
}

func TestReadAgentPromptFile_ExplicitTildeExpansion(t *testing.T) {
	promptBase := createPromptBaseForTest(t)

	path := filepath.Join(promptBase, "prompt.md")
	if err := os.WriteFile(path, []byte("P\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := readAgentPromptFile("~/.fish-agent-wrapper/prompts/prompt.md")
	if err != nil {
		t.Fatalf("readAgentPromptFile error: %v", err)
	}
	if got != "P" {
		t.Fatalf("got %q, want %q", got, "P")
	}
}

func TestReadAgentPromptFile_RestrictedAllowsClaudeDir(t *testing.T) {
	promptBase := createPromptBaseForTest(t)
	path := filepath.Join(promptBase, "prompt.md")
	if err := os.WriteFile(path, []byte("OK\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := readAgentPromptFile("~/.fish-agent-wrapper/prompts/prompt.md")
	if err != nil {
		t.Fatalf("readAgentPromptFile error: %v", err)
	}
	if got != "OK" {
		t.Fatalf("got %q, want %q", got, "OK")
	}
}

func TestReadAgentPromptFile_RestrictedRejectsOutsideClaudeDir(t *testing.T) {
	_ = createPromptBaseForTest(t)

	outsideDir := t.TempDir()
	path := filepath.Join(outsideDir, "prompt.md")
	if err := os.WriteFile(path, []byte("NO\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := readAgentPromptFile(path); err == nil {
		t.Fatalf("expected error for prompt file outside prompt base dir, got nil")
	}
}

func TestReadAgentPromptFile_RestrictedRejectsTraversal(t *testing.T) {
	promptBase := createPromptBaseForTest(t)

	secretDir := t.TempDir()
	path := filepath.Join(secretDir, "secret.md")
	if err := os.WriteFile(path, []byte("SECRET\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	traversal := filepath.Join(promptBase, "..", filepath.Base(path))
	if _, err := readAgentPromptFile(traversal); err == nil {
		t.Fatalf("expected traversal to be rejected, got nil")
	}
}

func TestReadAgentPromptFile_NotFound(t *testing.T) {
	promptBase := createPromptBaseForTest(t)

	_, err := readAgentPromptFile(filepath.Join(promptBase, "missing.md"))
	if err == nil || !os.IsNotExist(err) {
		t.Fatalf("expected not-exist error, got %v", err)
	}
}

func TestReadAgentPromptFile_PermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod-based permission test is not reliable on Windows")
	}

	promptBase := createPromptBaseForTest(t)

	path := filepath.Join(promptBase, "private.md")
	if err := os.WriteFile(path, []byte("PRIVATE\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}

	_, err := readAgentPromptFile(path)
	if err == nil {
		t.Fatalf("expected permission error, got nil")
	}
	if !os.IsPermission(err) && !strings.Contains(strings.ToLower(err.Error()), "permission") {
		t.Fatalf("expected permission denied, got: %v", err)
	}
}

func TestReadAgentPromptFile_RestrictedRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests are not reliable on Windows by default")
	}

	promptBase := createPromptBaseForTest(t)

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.md")
	if err := os.WriteFile(outsideFile, []byte("OUTSIDE\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	linkPath := filepath.Join(promptBase, "link.md")
	if err := os.Symlink(outsideFile, linkPath); err != nil {
		t.Skipf("Symlink not supported (%v)", err)
	}

	if _, err := readAgentPromptFile(linkPath); err == nil {
		t.Fatalf("expected symlink escape to be rejected, got nil")
	}
}
