---
name: code-dispatcher-flash
description: Execute code-dispatcher for multi-backend AI code tasks. Use when the user explicitly requests a specific backend (Codex, Claude, or Gemini), mentions code-dispatcher, or when a skill or command definition declares a dependency on this skill. Supports pluggable backends (codex/claude/gemini), parallel task execution with DAG scheduling, session resume, and structured output.
---

# code-dispatcher Usage

## Command Forms

**New task (HEREDOC, recommended):**
```bash
code-dispatcher --backend codex - [working_dir] <<'EOF'
<task content>
EOF
```

**New task (inline):**
```bash
code-dispatcher --backend codex "simple task" [working_dir]
```

**Resume (HEREDOC):**
```bash
code-dispatcher --backend codex resume <session_id> - <<'EOF'
<follow-up task>
EOF
```

**Resume (inline):**
```bash
code-dispatcher --backend claude resume <session_id> "follow-up task"
```

**Parallel:**
```bash
code-dispatcher --parallel --backend codex [--full-output] <<'EOF'
---TASK---
id: <unique_id>
[metadata fields]
---CONTENT---
<task content>

---TASK---
...
EOF
```

**Utility:**
```bash
code-dispatcher --help
code-dispatcher --cleanup          # Remove orphaned log files from temp dir
```

## Parameters

### Global Flags

- `--backend <backend>` (required): `codex | claude | gemini`. Also accepts `--backend=<backend>`.
- `--parallel`: Parallel task execution; reads task config from stdin.
- `--full-output`: Full per-task messages in parallel mode (default: summary). Only valid with `--parallel`.

### Positional Arguments

- `task`: Task text or `-` for stdin. Backend-native `@file` references work inside task text (handled by backends, not wrapper).
- `working_dir` (optional): Working directory. Default: current directory. Not valid with `--parallel` (use per-task `workdir` instead).

### Parallel Task Metadata Fields

Each `---TASK---` block supports these metadata fields (before `---CONTENT---`):

- `id` (required): Unique task identifier
- `backend`: Override global `--backend` for this task
- `workdir`: Working directory for this task (default: `.`)
- `dependencies`: Comma-separated task IDs this task depends on
- `session_id`: Resume an existing session (auto-sets mode to resume)

Validation: Lines without `:` are silently ignored. Unknown keys with `:` cause a parse error.

## Return Format

Single-task mode:
```text
Agent response text...

---
SESSION_ID: <uuid>
```

Use the wrapper-returned `SESSION_ID` as the source of truth for follow-up resume commands.

Parallel mode: Structured execution report (summary mode) or full per-task messages (`--full-output`).

## Backends

Supported: `codex | claude | gemini`. Per-task `backend` override in parallel mode allows mixed-backend execution.

## DAG Scheduling (Parallel Mode)
Tasks with `dependencies` run as a DAG:

```text
layer 0: task1      taskX       (concurrent)
           |          |
layer 1: task2      taskY       (concurrent)
             \      /
layer 2:      task3             (waits for both)
```

- Tasks in the same layer run concurrently
- Next layer starts only after current layer completes
- Failed dependency → dependent tasks skipped
- Invalid dependency IDs or cycles fail fast before execution starts

## Resume Session

- All backends support resume: `codex | claude | gemini`.
- Resume requires `--backend` and the `SESSION_ID` returned from a previous run.
- In parallel mode, any task block with `session_id` runs in resume mode automatically.
- Resume follows backend session context. If a different working directory is needed, start a new session.

## Exit Codes

- `0`: Success
- `1`: General error (missing args, no output, invalid config)
- `124`: Timeout
- `127`: Backend command not found in PATH
- `130`: Interrupted (Ctrl+C / SIGINT)
- Other: Passthrough from backend process

## Critical Rules

**NEVER kill code-dispatcher processes by default.** Long-running tasks (5–120 min) are normal.

1. **Check task status via log file:**
   ```bash
   # Log path is printed to stderr at startup
   # Format: $TMPDIR/code-dispatcher-{PID}[-{taskID}].log
   tail -f "$TMPDIR"/code-dispatcher-*.log
   ```

2. **Check process without killing:**
   ```bash
   ps aux | grep code-dispatcher | grep -v grep
   ```

3. **Set tool-call timeout by task complexity:**
   - Simple tasks: `600` s (10 min)
   - Normal tasks: `1800` s (30 min)
   - Complex Codex tasks: `7200` s (2 hr)
   - If the wait times out, do not kill the process; re-check logs/process and continue waiting.

   Pseudocode (adapt field names to your host runtime):
   ```text
   TaskOutput(task_id="<id>", block=true, timeout=1800)
   ```

**Why:** Killing wastes API costs and loses progress.

## Emergency Stop (User-Requested Only)

Kill/terminate is allowed **only when the user explicitly requests it**. Do not kill processes automatically because of long runtime or wait timeout. Name-based global cleanup (`pkill -x codex/claude/gemini`) is prohibited.

1. **Graceful stop wrapper first:**
   ```bash
   pgrep -fa code-dispatcher
   pkill -INT -f '(^|/)code-dispatcher( |$)'
   ```

2. **Escalate only if still running:**
   ```bash
   pkill -TERM -f '(^|/)code-dispatcher( |$)'
   sleep 2
   pkill -KILL -f '(^|/)code-dispatcher( |$)'
   ```

3. **Cleanup only descendants of target wrapper PID:**
   ```bash
   WRAPPER_PID=$(pgrep -n -f '(^|/)code-dispatcher( |$)')
   pkill -TERM -P "$WRAPPER_PID" 2>/dev/null || true
   sleep 2
   pkill -KILL -P "$WRAPPER_PID" 2>/dev/null || true
   ```

4. **Post-check:**
   ```bash
   pgrep -fa code-dispatcher
   pgrep -fa 'codex|claude|gemini'
   ```
