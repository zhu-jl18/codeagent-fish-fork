package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

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

const maxClaudeSettingsBytes = 1 << 20 // 1MB

type minimalClaudeSettings struct {
	Env map[string]string
}

// loadMinimalClaudeSettings 从 ~/.claude/settings.json 只提取安全的最小子集：
// - env: 只接受字符串类型的值
// - model: 只接受字符串类型的值
// 文件缺失/解析失败/超限都返回空。
func loadMinimalClaudeSettings() minimalClaudeSettings {
	claudeDir := resolveClaudeDir()
	if claudeDir == "" {
		return minimalClaudeSettings{}
	}

	settingPath := filepath.Join(claudeDir, "settings.json")
	info, err := os.Stat(settingPath)
	if err != nil || info.Size() > maxClaudeSettingsBytes {
		return minimalClaudeSettings{}
	}

	data, err := os.ReadFile(settingPath)
	if err != nil {
		return minimalClaudeSettings{}
	}

	var cfg struct {
		Env map[string]any `json:"env"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return minimalClaudeSettings{}
	}

	out := minimalClaudeSettings{}

	if len(cfg.Env) == 0 {
		return out
	}

	env := make(map[string]string, len(cfg.Env))
	for k, v := range cfg.Env {
		s, ok := v.(string)
		if !ok {
			continue
		}
		env[k] = s
	}
	if len(env) == 0 {
		return out
	}
	out.Env = env
	return out
}

// loadMinimalEnvSettings is kept for backwards tests; prefer loadMinimalClaudeSettings.
func loadMinimalEnvSettings() map[string]string {
	settings := loadMinimalClaudeSettings()
	if len(settings.Env) == 0 {
		return nil
	}
	return settings.Env
}

// loadGeminiEnv loads environment variables from ~/.gemini/.env
// Supports GEMINI_API_KEY, GEMINI_MODEL, GOOGLE_GEMINI_BASE_URL
// Also sets GEMINI_API_KEY_AUTH_MECHANISM=bearer for third-party API compatibility
func loadGeminiEnv() map[string]string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil
	}

	envPath := filepath.Join(home, ".gemini", ".env")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil
	}

	env := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if key != "" && value != "" {
			env[key] = value
		}
	}

	// Set bearer auth mechanism for third-party API compatibility
	if _, ok := env["GEMINI_API_KEY"]; ok {
		if _, hasAuth := env["GEMINI_API_KEY_AUTH_MECHANISM"]; !hasAuth {
			env["GEMINI_API_KEY_AUTH_MECHANISM"] = "bearer"
		}
	}

	if len(env) == 0 {
		return nil
	}
	return env
}

func buildClaudeArgs(cfg *Config, targetArg string) []string {
	if cfg == nil {
		return nil
	}
	args := []string{"-p"}
	// Default to skip permissions unless FISH_AGENT_WRAPPER_SKIP_PERMISSIONS=false
	if cfg.SkipPermissions || envFlagDefaultTrue("FISH_AGENT_WRAPPER_SKIP_PERMISSIONS") {
		args = append(args, "--dangerously-skip-permissions")
	}

	// Prevent infinite recursion: disable all setting sources (user, project, local)
	// This ensures a clean execution environment without CLAUDE.md or skills that would trigger fish-agent-wrapper
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

type AmpcodeBackend struct{}

func (AmpcodeBackend) Name() string { return "ampcode" }
func (AmpcodeBackend) Command() string {
	return "amp"
}
func (AmpcodeBackend) BuildArgs(cfg *Config, targetArg string) []string {
	return buildAmpcodeArgs(cfg, targetArg)
}

func resolveAmpcodeMode() string {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("FISH_AGENT_WRAPPER_AMPCODE_MODE")))
	switch raw {
	case "", "smart":
		return "smart"
	case "deep", "free", "rush":
		return raw
	default:
		logWarn("Invalid FISH_AGENT_WRAPPER_AMPCODE_MODE, fallback to smart")
		return "smart"
	}
}

func buildAmpcodeArgs(cfg *Config, targetArg string) []string {
	if cfg == nil {
		return nil
	}

	mode := resolveAmpcodeMode()
	args := []string{"--no-color", "--no-notifications"}

	if cfg.Mode == "resume" {
		if cfg.SessionID != "" {
			args = append(args, "threads", "continue", cfg.SessionID)
		}
	}

	if cfg.SkipPermissions || envFlagDefaultTrue("FISH_AGENT_WRAPPER_SKIP_PERMISSIONS") {
		args = append(args, "--dangerously-allow-all")
	}

	if targetArg == "-" {
		args = append(args, "--execute")
	} else {
		args = append(args, "--execute", targetArg)
	}

	args = append(args, "--stream-json", "--mode", mode)

	return args
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
