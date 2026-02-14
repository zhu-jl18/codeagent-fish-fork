package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestResumeConversation_SupportedBackends(t *testing.T) {
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
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			defer resetTestHooks()
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)

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

			first := runTaskWithContext(
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

			second := runTaskWithContext(
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
