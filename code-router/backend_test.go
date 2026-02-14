package main

import (
	"reflect"
	"testing"
)

func TestClaudeBuildArgs_ModesAndPermissions(t *testing.T) {
	backend := ClaudeBackend{}

	t.Run("new mode always includes skip-permissions", func(t *testing.T) {
		cfg := &Config{Mode: "new", WorkDir: "/repo"}
		got := backend.BuildArgs(cfg, "todo")
		want := []string{"-p", "--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose", "todo"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("stdin mode", func(t *testing.T) {
		cfg := &Config{Mode: "new", WorkDir: "/repo"}
		got := backend.BuildArgs(cfg, "-")
		want := []string{"-p", "--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose", "-"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("resume mode includes session id", func(t *testing.T) {
		cfg := &Config{Mode: "resume", SessionID: "sid-123", WorkDir: "/ignored"}
		got := backend.BuildArgs(cfg, "resume-task")
		want := []string{"-p", "--dangerously-skip-permissions", "-r", "sid-123", "--output-format", "stream-json", "--verbose", "resume-task"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("resume mode without session still returns base flags", func(t *testing.T) {
		cfg := &Config{Mode: "resume", WorkDir: "/ignored"}
		got := backend.BuildArgs(cfg, "follow-up")
		want := []string{"-p", "--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose", "follow-up"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("nil config returns nil", func(t *testing.T) {
		if backend.BuildArgs(nil, "ignored") != nil {
			t.Fatalf("nil config should return nil args")
		}
	})
}

func TestVariousBackendsBuildArgs(t *testing.T) {
	t.Run("gemini new mode defaults workdir", func(t *testing.T) {
		backend := GeminiBackend{}
		cfg := &Config{Mode: "new", WorkDir: "/workspace"}
		got := backend.BuildArgs(cfg, "task")
		want := []string{"-o", "stream-json", "-y", "task"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("gemini resume mode uses session id", func(t *testing.T) {
		backend := GeminiBackend{}
		cfg := &Config{Mode: "resume", SessionID: "sid-999"}
		got := backend.BuildArgs(cfg, "resume")
		want := []string{"-o", "stream-json", "-y", "-r", "sid-999", "resume"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("gemini resume mode without session omits identifier", func(t *testing.T) {
		backend := GeminiBackend{}
		cfg := &Config{Mode: "resume"}
		got := backend.BuildArgs(cfg, "resume")
		want := []string{"-o", "stream-json", "-y", "resume"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("gemini nil config returns nil", func(t *testing.T) {
		backend := GeminiBackend{}
		if backend.BuildArgs(nil, "ignored") != nil {
			t.Fatalf("nil config should return nil args")
		}
	})

	t.Run("gemini stdin mode uses -p flag", func(t *testing.T) {
		backend := GeminiBackend{}
		cfg := &Config{Mode: "new"}
		got := backend.BuildArgs(cfg, "-")
		want := []string{"-o", "stream-json", "-y", "-p", "-"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("codex build args always includes bypass flag", func(t *testing.T) {
		backend := CodexBackend{}
		cfg := &Config{Mode: "new", WorkDir: "/tmp"}
		got := backend.BuildArgs(cfg, "task")
		want := []string{"e", "--dangerously-bypass-approvals-and-sandbox", "--skip-git-repo-check", "-C", "/tmp", "--json", "task"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

}

func TestClaudeBuildArgs_BackendMetadata(t *testing.T) {
	tests := []struct {
		backend Backend
		name    string
		command string
	}{
		{backend: CodexBackend{}, name: "codex", command: "codex"},
		{backend: ClaudeBackend{}, name: "claude", command: "claude"},
		{backend: GeminiBackend{}, name: "gemini", command: "gemini"},
	}

	for _, tt := range tests {
		if got := tt.backend.Name(); got != tt.name {
			t.Fatalf("Name() = %s, want %s", got, tt.name)
		}
		if got := tt.backend.Command(); got != tt.command {
			t.Fatalf("Command() = %s, want %s", got, tt.command)
		}
	}
}

func TestRuntimeEnvForBackend(t *testing.T) {
	t.Run("returns nil when no runtime settings", func(t *testing.T) {
		setRuntimeSettingsForTest(map[string]string{})
		t.Cleanup(resetRuntimeSettingsForTest)
		if got := runtimeEnvForBackend("claude"); len(got) != 0 {
			t.Fatalf("got %v, want nil/empty", got)
		}
	})

	t.Run("filters wrapper control keys", func(t *testing.T) {
		setRuntimeSettingsForTest(map[string]string{
			"CODE_ROUTER_SKIP_PERMISSIONS": "false",
			"CODE_ROUTER_TIMEOUT":          "7200",
			"ANTHROPIC_API_KEY":            "secret",
			"FOO":                          "bar",
		})
		t.Cleanup(resetRuntimeSettingsForTest)

		got := runtimeEnvForBackend("claude")
		if got["ANTHROPIC_API_KEY"] != "secret" || got["FOO"] != "bar" {
			t.Fatalf("got %v, want ANTHROPIC_API_KEY/FOO", got)
		}
		if _, ok := got["CODE_ROUTER_TIMEOUT"]; ok {
			t.Fatalf("got %v, control key CODE_ROUTER_TIMEOUT should be filtered", got)
		}
		if _, ok := got["CODE_ROUTER_SKIP_PERMISSIONS"]; ok {
			t.Fatalf("got %v, control key CODE_ROUTER_SKIP_PERMISSIONS should be filtered", got)
		}
	})

	t.Run("gemini adds bearer auth fallback", func(t *testing.T) {
		setRuntimeSettingsForTest(map[string]string{
			"GEMINI_API_KEY": "k1",
		})
		t.Cleanup(resetRuntimeSettingsForTest)

		got := runtimeEnvForBackend("gemini")
		if got["GEMINI_API_KEY_AUTH_MECHANISM"] != "bearer" {
			t.Fatalf("got %v, want GEMINI_API_KEY_AUTH_MECHANISM=bearer", got)
		}
	})
}
