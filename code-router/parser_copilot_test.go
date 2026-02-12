package main

import (
	"strings"
	"testing"
)

func TestParseBackendStreamWithWarn_CopilotPlainTextFallback(t *testing.T) {
	var warnings []string
	warnFn := func(msg string) { warnings = append(warnings, msg) }

	message, threadID := parseBackendStreamWithWarn(strings.NewReader("line-1\nline-2"), "copilot", warnFn)
	if message != "line-1\nline-2\n" {
		t.Fatalf("message=%q, want %q", message, "line-1\nline-2\n")
	}
	if threadID != "" {
		t.Fatalf("threadID=%q, want empty", threadID)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings=%v, want none", warnings)
	}
}

func TestParseBackendStreamWithWarn_CopilotFallbackPreservesWhitespaceAndBlankLines(t *testing.T) {
	var warnings []string
	warnFn := func(msg string) { warnings = append(warnings, msg) }

	input := "  line-1\n\n\tline-2  \n"
	message, threadID := parseBackendStreamWithWarn(strings.NewReader(input), "copilot", warnFn)
	if message != input {
		t.Fatalf("message=%q, want %q", message, input)
	}
	if threadID != "" {
		t.Fatalf("threadID=%q, want empty", threadID)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings=%v, want none", warnings)
	}
}

func TestParseBackendStreamWithWarn_NonCopilotNoPlainTextFallback(t *testing.T) {
	var warnings []string
	warnFn := func(msg string) { warnings = append(warnings, msg) }

	message, threadID := parseBackendStreamWithWarn(strings.NewReader("line-1\nline-2"), "codex", warnFn)
	if message != "" || threadID != "" {
		t.Fatalf("expected empty output, got message=%q thread=%q", message, threadID)
	}
	if len(warnings) == 0 {
		t.Fatalf("expected warnings for non-json lines")
	}
}
