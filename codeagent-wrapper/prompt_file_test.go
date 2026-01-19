package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

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
	claudeDir := t.TempDir()
	t.Setenv("CODEAGENT_CLAUDE_DIR", claudeDir)

	path := filepath.Join(claudeDir, "prompt.md")
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
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	claudeDir := filepath.Join(home, ".claude")
	t.Setenv("CODEAGENT_CLAUDE_DIR", "~/.claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	path := filepath.Join(claudeDir, "prompt.md")
	if err := os.WriteFile(path, []byte("P\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := readAgentPromptFile("~/.claude/prompt.md")
	if err != nil {
		t.Fatalf("readAgentPromptFile error: %v", err)
	}
	if got != "P" {
		t.Fatalf("got %q, want %q", got, "P")
	}
}

func TestReadAgentPromptFile_RestrictedAllowsClaudeDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(claudeDir, "prompt.md")
	if err := os.WriteFile(path, []byte("OK\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := readAgentPromptFile("~/.claude/prompt.md")
	if err != nil {
		t.Fatalf("readAgentPromptFile error: %v", err)
	}
	if got != "OK" {
		t.Fatalf("got %q, want %q", got, "OK")
	}
}

func TestReadAgentPromptFile_RestrictedRejectsOutsideClaudeDir(t *testing.T) {
	claudeDir := t.TempDir()
	t.Setenv("CODEAGENT_CLAUDE_DIR", claudeDir)

	outsideDir := t.TempDir()
	path := filepath.Join(outsideDir, "prompt.md")
	if err := os.WriteFile(path, []byte("NO\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := readAgentPromptFile(path); err == nil {
		t.Fatalf("expected error for prompt file outside CODEAGENT_CLAUDE_DIR, got nil")
	}
}

func TestReadAgentPromptFile_RestrictedRejectsTraversal(t *testing.T) {
	claudeDir := t.TempDir()
	t.Setenv("CODEAGENT_CLAUDE_DIR", claudeDir)

	secretDir := t.TempDir()
	path := filepath.Join(secretDir, "secret.md")
	if err := os.WriteFile(path, []byte("SECRET\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	traversal := filepath.Join(claudeDir, "..", filepath.Base(path))
	if _, err := readAgentPromptFile(traversal); err == nil {
		t.Fatalf("expected traversal to be rejected, got nil")
	}
}

func TestReadAgentPromptFile_NotFound(t *testing.T) {
	claudeDir := t.TempDir()
	t.Setenv("CODEAGENT_CLAUDE_DIR", claudeDir)

	_, err := readAgentPromptFile(filepath.Join(claudeDir, "missing.md"))
	if err == nil || !os.IsNotExist(err) {
		t.Fatalf("expected not-exist error, got %v", err)
	}
}

func TestReadAgentPromptFile_PermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod-based permission test is not reliable on Windows")
	}

	claudeDir := t.TempDir()
	t.Setenv("CODEAGENT_CLAUDE_DIR", claudeDir)

	path := filepath.Join(claudeDir, "private.md")
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

	claudeDir := t.TempDir()
	t.Setenv("CODEAGENT_CLAUDE_DIR", claudeDir)

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.md")
	if err := os.WriteFile(outsideFile, []byte("OUTSIDE\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	linkPath := filepath.Join(claudeDir, "link.md")
	if err := os.Symlink(outsideFile, linkPath); err != nil {
		t.Skipf("Symlink not supported (%v)", err)
	}

	if _, err := readAgentPromptFile(linkPath); err == nil {
		t.Fatalf("expected symlink escape to be rejected, got nil")
	}
}
