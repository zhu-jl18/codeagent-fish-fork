package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestResumeConversation_AllBackends(t *testing.T) {
	type tc struct {
		name         string
		backend      Backend
		sessionID    string
		newOutput    string
		resumeOutput string
		checkNewArgs func(args []string) error
		checkResArgs func(args []string, sid string) error
	}

	cases := []tc{
		{
			name:      "codex",
			backend:   CodexBackend{},
			sessionID: "tid_codex",
			newOutput: `{"type":"thread.started","thread_id":"tid_codex"}` + "\n" +
				`{"type":"item.completed","item":{"type":"agent_message","text":"M1"}}` + "\n",
			resumeOutput: `{"type":"thread.started","thread_id":"tid_codex"}` + "\n" +
				`{"type":"item.completed","item":{"type":"agent_message","text":"M2"}}` + "\n",
			checkNewArgs: func(args []string) error {
				if strings.Contains(strings.Join(args, " "), " resume ") {
					return fmt.Errorf("unexpected resume in args: %v", args)
				}
				return nil
			},
			checkResArgs: func(args []string, sid string) error {
				joined := strings.Join(args, " ")
				if !strings.Contains(joined, " resume "+sid+" ") {
					return fmt.Errorf("missing resume %s in args: %v", sid, args)
				}
				return nil
			},
		},
		{
			name:         "claude",
			backend:      ClaudeBackend{},
			sessionID:    "sid_claude",
			newOutput:    `{"type":"result","session_id":"sid_claude","result":"M1"}` + "\n",
			resumeOutput: `{"type":"result","session_id":"sid_claude","result":"M2"}` + "\n",
			checkNewArgs: func(args []string) error {
				if strings.Contains(strings.Join(args, " "), " -r ") {
					return fmt.Errorf("unexpected -r in args: %v", args)
				}
				return nil
			},
			checkResArgs: func(args []string, sid string) error {
				joined := strings.Join(args, " ")
				if !strings.Contains(joined, " -r "+sid+" ") {
					return fmt.Errorf("missing -r %s in args: %v", sid, args)
				}
				return nil
			},
		},
		{
			name:         "gemini",
			backend:      GeminiBackend{},
			sessionID:    "sid_gemini",
			newOutput:    `{"type":"result","session_id":"sid_gemini","status":"success","content":"M1"}` + "\n",
			resumeOutput: `{"type":"result","session_id":"sid_gemini","status":"success","content":"M2"}` + "\n",
			checkNewArgs: func(args []string) error {
				if strings.Contains(strings.Join(args, " "), " -r ") {
					return fmt.Errorf("unexpected -r in args: %v", args)
				}
				return nil
			},
			checkResArgs: func(args []string, sid string) error {
				joined := strings.Join(args, " ")
				if !strings.Contains(joined, " -r "+sid+" ") {
					return fmt.Errorf("missing -r %s in args: %v", sid, args)
				}
				return nil
			},
		},
		{
			name:      "ampcode",
			backend:   AmpcodeBackend{},
			sessionID: "T-ampcode",
			newOutput: `{"type":"assistant","session_id":"T-ampcode","message":{"content":[{"type":"text","text":"M1"}]}}` + "\n",
			resumeOutput: `{"type":"assistant","session_id":"T-ampcode","message":{"content":[{"type":"text","text":"M2"}]}}` + "\n" +
				`{"type":"done","session_id":"T-ampcode"}` + "\n",
			checkNewArgs: func(args []string) error {
				joined := " " + strings.Join(args, " ") + " "
				if strings.Contains(joined, " threads continue ") {
					return fmt.Errorf("unexpected threads continue in args: %v", args)
				}
				return nil
			},
			checkResArgs: func(args []string, sid string) error {
				joined := " " + strings.Join(args, " ") + " "
				if !strings.Contains(joined, " threads continue "+sid+" ") {
					return fmt.Errorf("missing threads continue %s in args: %v", sid, args)
				}
				return nil
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			defer resetTestHooks()
			t.Setenv("CODEX_BYPASS_SANDBOX", "false")
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			t.Setenv("FISH_AGENT_WRAPPER_CLAUDE_DIR", t.TempDir())

			var calls int
			newCommandRunner = func(ctx context.Context, name string, args ...string) commandRunner {
				calls++
				if name != tt.backend.Command() {
					t.Fatalf("command name=%q, want %q", name, tt.backend.Command())
				}
				if calls == 1 {
					if err := tt.checkNewArgs(args); err != nil {
						t.Fatalf("new args check failed: %v", err)
					}
					return newFakeCmd(fakeCmdConfig{StdoutPlan: []fakeStdoutEvent{{Data: tt.newOutput}}})
				}
				if calls == 2 {
					if err := tt.checkResArgs(args, tt.sessionID); err != nil {
						t.Fatalf("resume args check failed: %v", err)
					}
					return newFakeCmd(fakeCmdConfig{StdoutPlan: []fakeStdoutEvent{{Data: tt.resumeOutput}}})
				}
				t.Fatalf("unexpected extra call %d (args=%v)", calls, args)
				return newFakeCmd(fakeCmdConfig{})
			}

			first := runCodexTaskWithContext(
				context.Background(),
				TaskSpec{ID: "new", Task: "hello", WorkDir: ".", Mode: "new"},
				tt.backend,
				nil,
				false,
				true,
				5,
			)
			if first.ExitCode != 0 || first.Error != "" {
				t.Fatalf("new failed: exit=%d err=%q res=%+v", first.ExitCode, first.Error, first)
			}
			if first.SessionID != tt.sessionID {
				t.Fatalf("new session=%q, want %q (res=%+v)", first.SessionID, tt.sessionID, first)
			}

			second := runCodexTaskWithContext(
				context.Background(),
				TaskSpec{ID: "resume", Task: "follow-up", WorkDir: ".", Mode: "resume", SessionID: first.SessionID},
				tt.backend,
				nil,
				false,
				true,
				5,
			)
			if second.ExitCode != 0 || second.Error != "" {
				t.Fatalf("resume failed: exit=%d err=%q res=%+v", second.ExitCode, second.Error, second)
			}
			if second.SessionID != tt.sessionID {
				t.Fatalf("resume session=%q, want %q (res=%+v)", second.SessionID, tt.sessionID, second)
			}
		})
	}
}

func TestRunParallel_AllBackends_NewMode(t *testing.T) {
	defer resetTestHooks()
	t.Setenv("CODEX_BYPASS_SANDBOX", "false")
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("FISH_AGENT_WRAPPER_CLAUDE_DIR", t.TempDir())

	os.Args = []string{"fish-agent-wrapper", "--parallel"}
	stdinReader = bytes.NewReader([]byte(
		`---TASK---
id: codex
backend: codex
---CONTENT---
hello-codex
---TASK---
id: claude
backend: claude
---CONTENT---
hello-claude
---TASK---
id: gemini
backend: gemini
---CONTENT---
hello-gemini
---TASK---
id: amp
backend: ampcode
---CONTENT---
hello-amp
`,
	))

	var (
		mu      sync.Mutex
		runErr  error
		seenCmd = make(map[string]bool)
	)

	newCommandRunner = func(ctx context.Context, name string, args ...string) commandRunner {
		joined := " " + strings.Join(args, " ") + " "
		mu.Lock()
		defer mu.Unlock()
		if runErr != nil {
			return newFakeCmd(fakeCmdConfig{})
		}
		seenCmd[name] = true
		switch name {
		case "codex":
			if strings.Contains(joined, " resume ") {
				runErr = fmt.Errorf("codex new unexpectedly contains resume: %v", args)
			}
			out := `{"type":"thread.started","thread_id":"tid_codex"}` + "\n" +
				`{"type":"item.completed","item":{"type":"agent_message","text":"OK"}}` + "\n"
			return newFakeCmd(fakeCmdConfig{StdoutPlan: []fakeStdoutEvent{{Data: out}}})
		case "claude":
			if strings.Contains(joined, " -r ") {
				runErr = fmt.Errorf("claude new unexpectedly contains -r: %v", args)
			}
			out := `{"type":"result","session_id":"sid_claude","result":"OK"}` + "\n"
			return newFakeCmd(fakeCmdConfig{StdoutPlan: []fakeStdoutEvent{{Data: out}}})
		case "gemini":
			if strings.Contains(joined, " -r ") {
				runErr = fmt.Errorf("gemini new unexpectedly contains -r: %v", args)
			}
			out := `{"type":"result","session_id":"sid_gemini","status":"success","content":"OK"}` + "\n"
			return newFakeCmd(fakeCmdConfig{StdoutPlan: []fakeStdoutEvent{{Data: out}}})
		case "amp":
			if strings.Contains(joined, " threads continue ") {
				runErr = fmt.Errorf("ampcode new unexpectedly contains threads continue: %v", args)
			}
			out := `{"type":"assistant","session_id":"T-amp-new","message":{"content":[{"type":"text","text":"OK"}]}}` + "\n" +
				`{"type":"done","session_id":"T-amp-new"}` + "\n"
			return newFakeCmd(fakeCmdConfig{StdoutPlan: []fakeStdoutEvent{{Data: out}}})
		default:
			runErr = fmt.Errorf("unexpected command: %s (args=%v)", name, args)
			return newFakeCmd(fakeCmdConfig{})
		}
	}

	var exitCode int
	out := captureOutput(t, func() { exitCode = run() })

	if exitCode != 0 {
		t.Fatalf("run() exit=%d, want 0 (out=%q)", exitCode, out)
	}
	if runErr != nil {
		t.Fatalf("runner error: %v", runErr)
	}
	for _, cmd := range []string{"codex", "claude", "gemini", "amp"} {
		if !seenCmd[cmd] {
			t.Fatalf("did not run backend %q", cmd)
		}
	}

	payload := parseIntegrationOutput(t, out)
	if payload.Summary.Total != 4 || payload.Summary.Failed != 0 || payload.Summary.Success != 4 {
		t.Fatalf("unexpected summary: %+v (out=%q)", payload.Summary, out)
	}
}

func TestRunParallel_AllBackends_ResumeMode(t *testing.T) {
	defer resetTestHooks()
	t.Setenv("CODEX_BYPASS_SANDBOX", "false")
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("FISH_AGENT_WRAPPER_CLAUDE_DIR", t.TempDir())

	os.Args = []string{"fish-agent-wrapper", "--parallel"}
	stdinReader = bytes.NewReader([]byte(
		`---TASK---
id: codex
backend: codex
session_id: tid_codex
---CONTENT---
resume-codex
---TASK---
id: claude
backend: claude
session_id: sid_claude
---CONTENT---
resume-claude
---TASK---
id: gemini
backend: gemini
session_id: sid_gemini
---CONTENT---
resume-gemini
---TASK---
id: amp
backend: ampcode
session_id: T-amp-resume
---CONTENT---
resume-amp
`,
	))

	var (
		mu      sync.Mutex
		runErr  error
		seenCmd = make(map[string]bool)
	)

	newCommandRunner = func(ctx context.Context, name string, args ...string) commandRunner {
		joined := " " + strings.Join(args, " ") + " "
		mu.Lock()
		defer mu.Unlock()
		if runErr != nil {
			return newFakeCmd(fakeCmdConfig{})
		}
		seenCmd[name] = true
		switch name {
		case "codex":
			if !strings.Contains(joined, " resume tid_codex ") {
				runErr = fmt.Errorf("codex resume missing tid_codex: %v", args)
			}
			out := `{"type":"thread.started","thread_id":"tid_codex"}` + "\n" +
				`{"type":"item.completed","item":{"type":"agent_message","text":"OK"}}` + "\n"
			return newFakeCmd(fakeCmdConfig{StdoutPlan: []fakeStdoutEvent{{Data: out}}})
		case "claude":
			if !strings.Contains(joined, " -r sid_claude ") {
				runErr = fmt.Errorf("claude resume missing -r sid_claude: %v", args)
			}
			out := `{"type":"result","session_id":"sid_claude","result":"OK"}` + "\n"
			return newFakeCmd(fakeCmdConfig{StdoutPlan: []fakeStdoutEvent{{Data: out}}})
		case "gemini":
			if !strings.Contains(joined, " -r sid_gemini ") {
				runErr = fmt.Errorf("gemini resume missing -r sid_gemini: %v", args)
			}
			out := `{"type":"result","session_id":"sid_gemini","status":"success","content":"OK"}` + "\n"
			return newFakeCmd(fakeCmdConfig{StdoutPlan: []fakeStdoutEvent{{Data: out}}})
		case "amp":
			if !strings.Contains(joined, " threads continue T-amp-resume ") {
				runErr = fmt.Errorf("ampcode resume missing threads continue T-amp-resume: %v", args)
			}
			out := `{"type":"assistant","session_id":"T-amp-resume","message":{"content":[{"type":"text","text":"OK"}]}}` + "\n" +
				`{"type":"done","session_id":"T-amp-resume"}` + "\n"
			return newFakeCmd(fakeCmdConfig{StdoutPlan: []fakeStdoutEvent{{Data: out}}})
		default:
			runErr = fmt.Errorf("unexpected command: %s (args=%v)", name, args)
			return newFakeCmd(fakeCmdConfig{})
		}
	}

	var exitCode int
	out := captureOutput(t, func() { exitCode = run() })

	if exitCode != 0 {
		t.Fatalf("run() exit=%d, want 0 (out=%q)", exitCode, out)
	}
	if runErr != nil {
		t.Fatalf("runner error: %v", runErr)
	}
	for _, cmd := range []string{"codex", "claude", "gemini", "amp"} {
		if !seenCmd[cmd] {
			t.Fatalf("did not run backend %q", cmd)
		}
	}

	payload := parseIntegrationOutput(t, out)
	if payload.Summary.Total != 4 || payload.Summary.Failed != 0 || payload.Summary.Success != 4 {
		t.Fatalf("unexpected summary: %+v (out=%q)", payload.Summary, out)
	}
}
