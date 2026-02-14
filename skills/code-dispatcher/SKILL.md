---
name: code-dispatcher
description: Execute code-dispatcher for multi-backend AI code tasks. Supports Codex, Claude, Gemini, with file references (@syntax) and structured output.
---

# code-dispatcher Integration

## Overview

Execute code-dispatcher commands with pluggable AI backends (Codex, Claude, Gemini). Supports file references via `@` syntax, parallel task execution with backend selection, and configurable security controls.

## When to Use

When the user explicitly requests a specific backend (Codex, Claude, or Gemini), mentions code-dispatcher, or when a skill or command definition explicitly declares a dependency on this skill.

Applicable scenarios include but are not limited to:
- Complex code analysis requiring deep understanding
- Large-scale refactoring across multiple files
- Automated code generation with backend selection


## Typical Usage for One-Round/New Tasks:

**1) Standard invocation: HEREDOC syntax (recommended)**
```bash
code-dispatcher --backend codex - [working_dir] <<'EOF'
<task content here>
EOF
```

**2) Single-line tasks (no heredoc)**
```bash
code-dispatcher --backend codex "simple task" [working_dir]
code-dispatcher --backend claude "simple task" [working_dir]
code-dispatcher --backend gemini "simple task" [working_dir]
```

## Common Parameters

- Command notation and positional order
  - `[]` means optional. Do not type brackets literally.
  - Inline task: `code-dispatcher --backend <backend> "<task>" [working_dir]` — task text is passed in command args; best for short one-line prompts.
  - Stdin task: `code-dispatcher --backend <backend> - [working_dir]` — task text is read from stdin (`<<'EOF'`/pipe); best for multi-line or complex content.
  - These two forms are for new-task commands only.
  - Resume uses its own forms: inline `code-dispatcher --backend <backend> resume <session_id> "<follow-up task>"` and stdin `code-dispatcher --backend <backend> resume <session_id> -`.

- `--backend <backend>` (required)
  - Select backend explicitly: `codex | claude | gemini`.
  - Must be present in both new and resume modes.
  - In parallel mode, this is the global default backend.
  - If a task block defines `backend`, it overrides the global default for that task.

- `task` (required)
  - Task description for the backend.
  - Supports inline text or stdin marker `-`.
  - Supports `@file` references (backend-native feature, not processed by wrapper).

- `working_dir` (optional)
  - Working directory for new task execution.
  - Omit it to use the current directory.
  - Typical values: `.`, `./subdir`, `/absolute/path`.
  - In resume mode, do not append `working_dir`; resume follows backend session context.

- Output modes (parallel execution only)
  - **Summary (default)**: Structured report with changes, output, verification, and review summary.
  - **Full (`--full-output`)**: Complete task messages. Use only when debugging specific failures.
  - Scope: `--full-output` is valid only with `--parallel`; single-task mode does not support this flag.
  - Backend behavior: mode selection is wrapper-level and works the same for `codex | claude | gemini`.

## Return Format:

```
Agent response text here...

---
SESSION_ID: 019a7247-ac9d-71f3-89e2-a823dbd8fd14
```

## Backends Selection Guide

**Note**: This backends selection guide applies only when the user has not explicitly requested a specific backend. If the user specifies a backend, always follow the user's instructions.


Quick look at the differences between backends:

| Backend | Command | Description | Best For |
|---------|---------|-------------|----------|
| codex | `--backend codex` | OpenAI Codex (default) | Code analysis, complex development, debugging |
| claude | `--backend claude` | Anthropic Claude | Quick fixes, documentation, prompts |
| gemini | `--backend gemini` | Google Gemini | UI/UX prototyping |

For detailed guidance:

**Codex**:
- Deep code understanding and complex logic implementation
- Large-scale refactoring with precise dependency tracking
- Algorithm optimization and performance tuning
- Example: "Analyze the call graph of @src/core and refactor the module dependency structure"

**Claude**:
- Quick feature implementation with clear requirements
- Technical documentation, API specs, README generation
- Professional prompt engineering (e.g., product requirements, design specs)
- Example: "Generate a comprehensive README for @package.json with installation, usage, and API docs"

**Gemini**:
- UI component scaffolding and layout prototyping
- Design system implementation with style consistency
- Interactive element generation with accessibility support
- Example: "Create a responsive dashboard layout with sidebar navigation and data visualization cards"

A Typical Backend Switching Example:
- Start with Codex for analysis, switch to Claude for documentation, then Gemini for UI implementation.
- Use per-task backend selection in parallel mode to optimize for each task's strengths

## Resume Session

Supported backends all support resume mode: `codex | claude | gemini`.

**1) Standard resume (HEREDOC)**
```bash
code-dispatcher --backend codex resume <session_id> - <<'EOF'
<follow-up task>
EOF
```

**2) Single-line resume (no heredoc)**
```bash
code-dispatcher --backend claude resume <session_id> "follow-up task"
```

**3) Parallel resume (supported)**
```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: resume-a
backend: claude
session_id: sid_claude_1
---CONTENT---
follow-up for claude session

EOF
```

In parallel mode, any task that provides `session_id` runs in resume mode.

Resume mode relies on backend session context.
- Do not append `[working_dir]` in resume commands.
- If you need a different directory, start a new session instead of resume.

 Resume identifier contract:
- Use the wrapper-returned `SESSION_ID` as the source of truth for follow-up resume commands.
- Standard form: `code-dispatcher --backend <backend> resume <SESSION_ID> ...`.


## Parallel Mode (`--parallel`)

`--parallel` is the execution entry for submitting multiple tasks in one run.

- Each task must define a unique `id`.
- `--backend` is a required global fallback; tasks without `backend` use it. Usually set to `codex`.
- `backend` inside a task block overrides the global fallback for that task.
- Tasks without `dependencies` are independent and can run concurrently.

**1) Basic parallel (independent tasks, global backend fallback)**
```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: scan-api
---CONTENT---
scan @src/api and summarize route structure

---TASK---
id: scan-db
---CONTENT---
scan @src/db and summarize data access patterns
EOF
```

**2) Parallel with per-task backend override (mixed backends)**
```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: analyze
---CONTENT---
analyze code structure

---TASK---
id: document
backend: claude
---CONTENT---
write technical notes based on the analysis

---TASK---
id: ui-draft
backend: gemini
---CONTENT---
draft UI layout from the same requirements
EOF
```

Output styles in parallel mode:

**1) Summary mode (default, no flag)**
```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: t1
---CONTENT---
analyze @src and summarize architecture changes
EOF
```

**2) Full mode (`--full-output`)**, mainly for debugging failures or when full per-task messages are required.
```bash
code-dispatcher --parallel --backend codex --full-output <<'EOF'
---TASK---
id: t1
---CONTENT---
analyze @src and summarize architecture changes
EOF
```

## DAG Scheduling (inside `--parallel`)

When a task declares `dependencies`, `--parallel` runs tasks as a DAG scheduler.

- `dependencies: a, b` means this task waits for tasks `a` and `b` to succeed.
- Tasks in the same DAG layer run concurrently; the next layer starts only after the current layer finishes.
- If a dependency fails, dependent tasks are skipped.
- Invalid dependency IDs or dependency cycles fail fast before execution starts.

ASCII execution model:
```text
layer 0: task1      taskX
           |          |
layer 1: task2      taskY
             \      /
layer 2:      task3
```

**1) Dependency scheduling (DAG with mixed backends)**
```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: task1
workdir: /path/to/dir
---CONTENT---
analyze code structure

---TASK---
id: task2
backend: claude
dependencies: task1
---CONTENT---
design architecture based on task1 analysis

---TASK---
id: task3
backend: gemini
dependencies: task2
---CONTENT---
generate implementation code
EOF
```

**2) Minimal DAG example (annotated)**
```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: prep
# uses global backend codex
---CONTENT---
scan @src and list key modules

---TASK---
id: plan
backend: claude
# overrides global backend for this task
dependencies: prep
---CONTENT---
write implementation plan based on prep output
EOF
```

## Runtime Config and Patterns

- Runtime environment and approval policy are pre-configured by the human/operator.
- Do not inject, override, or document environment variable values in this skill.
- Treat wrapper-internal timeout as a very long fallback configured by the operator.
- Control actual waiting budget via the host tool-call timeout.
- Timeout tiers for tool-call invocations:
  - Simple tasks: `600` s (10 minutes) minimum.
  - Normal tasks: `1800` s (30 minutes) recommended default.
  - Complex Codex tasks: `7200` s (2 hours).
- Do not use short timeouts like `300` s (5 minutes) for normal or complex tasks.

Invocation Pattern:

**Single Task**:
```
Host-agnostic tool-call template (field names vary by runtime):
- command payload (`command` or `cmd`):
  code-dispatcher --backend <backend> - [working_dir] <<'EOF'
  <task content>
  EOF
- timeout field (`timeout` / `timeout_ms` / equivalent): choose by tier (`600` / `1800` / `7200`)
- description field: optional

Field names depend on the host tool schema.
Timeout policy: always set explicit timeout by task complexity; do not rely on implicit defaults.

Note: `--backend` is required; supported values: `codex | claude | gemini`
```

**Parallel Tasks**:
```
Host-agnostic tool-call template (field names vary by runtime):
- command payload (`command` or `cmd`):
  code-dispatcher --parallel --backend <backend> <<'EOF'
  ---TASK---
  id: task_id
  backend: <backend>  # Optional, overrides global
  workdir: /path
  dependencies: dep1, dep2
  ---CONTENT---
  task content
  EOF
- timeout field (`timeout` / `timeout_ms` / equivalent): choose by tier (`600` / `1800` / `7200`)
- description field: optional

Field names depend on the host tool schema.
Timeout policy: always set explicit timeout by task complexity; do not rely on implicit defaults.

Note: Global --backend is required; per-task backend is optional
```

## Critical Rules

**NEVER kill code-dispatcher processes by default.** Long-running tasks are normal. Instead:

1. **Check task status via log file**:
   ```bash
   # Log path is printed to stderr at startup
   # Format: $TMPDIR/code-dispatcher-{PID}[-{taskID}].log
   tail -f "$TMPDIR"/code-dispatcher-*.log
   ```

2. **Wait with tiered timeout (host-runtime API)**:
  - Use the host runtime's blocking wait API (for example: TaskOutput/wait-result equivalents).
  - Choose timeout by complexity:
    - Simple: `600` s (10m)
    - Normal: `1800` s (30m)
    - Complex Codex: `7200` s (2h)
  - If the wait call times out, do not kill the process; re-check logs/process and continue waiting.

  - Pseudocode (adapt field names to your host runtime):
   ```text
   TaskOutput(task_id="<id>", block=true, timeout=600)
   TaskOutput(task_id="<id>", block=true, timeout=1800)
   TaskOutput(task_id="<id>", block=true, timeout=7200)
   ```

3. **Check process without killing**:
   ```bash
   ps aux | grep code-dispatcher | grep -v grep
   ```

**Why:** code-dispatcher tasks often take 5-120 minutes. Killing them wastes API costs and loses progress.

## Emergency Stop (User-Requested Only)

- Hard rule: kill/terminate is allowed **only when the user explicitly requests it**.
- Do not kill processes automatically because of long runtime or wait timeout.
- Use staged termination and stop escalation as soon as processes exit.
- Name-based global cleanup (`pkill -x codex/claude/gemini`) is prohibited.

1. **Graceful stop wrapper first**:
   ```bash
   # Inspect running wrapper processes
   pgrep -fa code-dispatcher

   # Soft interrupt first
   pkill -INT -f '(^|/)code-dispatcher( |$)'
   ```

2. **Escalate only if still running**:
   ```bash
   pkill -TERM -f '(^|/)code-dispatcher( |$)'
   sleep 2
   pkill -KILL -f '(^|/)code-dispatcher( |$)'
   ```

3. **Cleanup only descendants of the target code-dispatcher PID (safe default)**:
   ```bash
   # Pick target code-dispatcher PID first (example: newest one)
   DISPATCHER_PID=$(pgrep -n -f '(^|/)code-dispatcher( |$)')

   # TERM direct children of this dispatcher only
   pkill -TERM -P "$DISPATCHER_PID" 2>/dev/null || true
   sleep 2

   # If still present, escalate to KILL for direct children only
   pkill -KILL -P "$DISPATCHER_PID" 2>/dev/null || true
   ```

4. **Post-check**:
   ```bash
   pgrep -fa code-dispatcher
   pgrep -fa 'codex|claude|gemini'
   ```
