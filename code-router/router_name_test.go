package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCurrentWrapperNameFallsBackToExecutable(t *testing.T) {
	defer resetTestHooks()

	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "code-router")
	if err := os.WriteFile(execPath, []byte("#!/bin/true\n"), 0o755); err != nil {
		t.Fatalf("failed to write fake binary: %v", err)
	}

	os.Args = []string{filepath.Join(tempDir, "custom-name")}
	executablePathFn = func() (string, error) {
		return execPath, nil
	}

	if got := currentRouterName(); got != defaultRouterName {
		t.Fatalf("currentRouterName() = %q, want %q", got, defaultRouterName)
	}
}
