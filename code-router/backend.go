package main

import "strings"

// Backend defines the contract for invoking different AI CLI backends.
// Each backend is responsible for supplying the executable command and
// building the argument list based on the wrapper config.
type Backend interface {
	Name() string
	BuildArgs(cfg *Config, targetArg string) []string
	Command() string
}

type CodexBackend struct{}

func (CodexBackend) Name() string { return "codex" }
func (CodexBackend) Command() string {
	return "codex"
}
func (CodexBackend) BuildArgs(cfg *Config, targetArg string) []string {
	return buildCodexArgs(cfg, targetArg)
}

type ClaudeBackend struct{}

func (ClaudeBackend) Name() string { return "claude" }
func (ClaudeBackend) Command() string {
	return "claude"
}
func (ClaudeBackend) BuildArgs(cfg *Config, targetArg string) []string {
	return buildClaudeArgs(cfg, targetArg)
}

func runtimeEnvForBackend(backendName string) map[string]string {
	env := runtimeInjectedEnv()
	if len(env) == 0 {
		return nil
	}

	if normalizeBackendName(backendName) == "gemini" {
		if _, ok := env["GEMINI_API_KEY"]; ok {
			if _, hasAuth := env["GEMINI_API_KEY_AUTH_MECHANISM"]; !hasAuth {
				env["GEMINI_API_KEY_AUTH_MECHANISM"] = "bearer"
			}
		}
	}

	return env
}

func normalizeBackendName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func backendSupportsStdinPrompt(name string) bool {
	return normalizeBackendName(name) != "copilot"
}

func backendAllowsPlainTextOutput(name string) bool {
	return normalizeBackendName(name) == "copilot"
}

func buildClaudeArgs(cfg *Config, targetArg string) []string {
	if cfg == nil {
		return nil
	}
	args := []string{"-p"}
	// Default to skip permissions unless CODE_ROUTER_SKIP_PERMISSIONS=false
	if cfg.SkipPermissions || envFlagDefaultTrue("CODE_ROUTER_SKIP_PERMISSIONS") {
		args = append(args, "--dangerously-skip-permissions")
	}

	// Prevent infinite recursion: disable all setting sources (user, project, local)
	// This ensures a clean execution environment without CLAUDE.md or skills that would trigger code-router
	args = append(args, "--setting-sources", "")

	if cfg.Mode == "resume" {
		if cfg.SessionID != "" {
			// Claude CLI uses -r <session_id> for resume.
			args = append(args, "-r", cfg.SessionID)
		}
	}
	// Note: claude CLI doesn't support -C flag; workdir set via cmd.Dir

	args = append(args, "--output-format", "stream-json", "--verbose", targetArg)

	return args
}

type GeminiBackend struct{}

func (GeminiBackend) Name() string { return "gemini" }
func (GeminiBackend) Command() string {
	return "gemini"
}
func (GeminiBackend) BuildArgs(cfg *Config, targetArg string) []string {
	return buildGeminiArgs(cfg, targetArg)
}

func buildGeminiArgs(cfg *Config, targetArg string) []string {
	if cfg == nil {
		return nil
	}
	args := []string{"-o", "stream-json", "-y"}

	if cfg.Mode == "resume" {
		if cfg.SessionID != "" {
			args = append(args, "-r", cfg.SessionID)
		}
	}
	// Note: gemini CLI doesn't support -C flag; workdir set via cmd.Dir

	// Use positional argument instead of deprecated -p flag
	// For stdin mode ("-"), use -p to read from stdin
	if targetArg == "-" {
		args = append(args, "-p", targetArg)
	} else {
		args = append(args, targetArg)
	}

	return args
}

type CopilotBackend struct{}

func (CopilotBackend) Name() string { return "copilot" }
func (CopilotBackend) Command() string {
	return "copilot"
}
func (CopilotBackend) BuildArgs(cfg *Config, targetArg string) []string {
	return buildCopilotArgs(cfg, targetArg)
}

func buildCopilotArgs(cfg *Config, targetArg string) []string {
	if cfg == nil {
		return nil
	}
	args := make([]string, 0, 8)

	if cfg.SkipPermissions || envFlagDefaultTrue("CODE_ROUTER_SKIP_PERMISSIONS") {
		args = append(args, "--allow-all")
	}

	args = append(args, "--no-custom-instructions")

	if cfg.Mode == "resume" {
		if cfg.SessionID != "" {
			args = append(args, "--resume", cfg.SessionID)
		} else {
			args = append(args, "--continue")
		}
	}

	args = append(args, "--silent", "-p", targetArg)

	return args
}
