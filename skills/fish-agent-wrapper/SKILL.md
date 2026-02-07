---
name: fish-agent-wrapper
description: Execute fish-agent-wrapper for multi-backend AI code tasks. Supports Codex, Claude, Gemini, Ampcode, with file references (@syntax) and structured output.
---

# fish-agent-wrapper Integration

## Overview

Execute fish-agent-wrapper commands with pluggable AI backends(Codex, Claude, Gemini, Ampcode). Supports file references via `@` syntax, parallel task execution with backend selection, and configurable security controls.

## When to Use

- Complex code analysis requiring deep understanding
- Large-scale refactoring across multiple files
- Automated code generation with backend selection

## Usage

**HEREDOC syntax** (recommended):
```bash
fish-agent-wrapper --backend codex - [working_dir] <<'EOF'
<task content here>
EOF
```

**With backend selection**:
```bash
fish-agent-wrapper --backend claude - . <<'EOF'
<task content here>
EOF
```

**Simple tasks**:
```bash
fish-agent-wrapper --backend codex "simple task" [working_dir]
fish-agent-wrapper --backend gemini "simple task" [working_dir]
fish-agent-wrapper --backend ampcode "simple task" [working_dir]
```

## Backends

| Backend | Command | Description | Best For |
|---------|---------|-------------|----------|
| codex | `--backend codex` | OpenAI Codex (default) | Code analysis, complex development |
| claude | `--backend claude` | Anthropic Claude | Quick fixes, documentation, prompts |
| gemini | `--backend gemini` | Google Gemini | UI/UX prototyping |
| ampcode | `--backend ampcode` | Amp CLI backend | Review tasks |

### Backend Selection Guide

**Codex** (default):
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

**Ampcode**:
- Standard backend option for tasks that specify `--backend ampcode`
- Example: "Review @.claude/specs/auth/dev-plan.md"

**Backend Switching**:
- Start with Codex for analysis, switch to Claude for documentation, then Gemini for UI implementation
- Use per-task backend selection in parallel mode to optimize for each task's strengths

## Parameters

- `task` (required): Task description, supports `@file` references
- `working_dir` (optional): Working directory (default: current)
- `--backend` (required): Select AI backend (codex/claude/gemini/ampcode)

## Return Format

```
Agent response text here...

---
SESSION_ID: 019a7247-ac9d-71f3-89e2-a823dbd8fd14
```

## Resume Session

```bash
# Resume with codex backend
fish-agent-wrapper --backend codex resume <session_id> - <<'EOF'
<follow-up task>
EOF

# Resume with specific backend
fish-agent-wrapper --backend claude resume <session_id> - <<'EOF'
<follow-up task>
EOF

# Resume with ampcode backend
fish-agent-wrapper --backend ampcode resume <session_id> - <<'EOF'
<follow-up task>
EOF
```

## Parallel Execution

**Default (summary mode - context-efficient):**
```bash
fish-agent-wrapper --parallel <<'EOF'
---TASK---
id: task1
backend: codex
workdir: /path/to/dir
---CONTENT---
task content
---TASK---
id: task2
dependencies: task1
---CONTENT---
dependent task
EOF
```

**Full output mode (for debugging):**
```bash
fish-agent-wrapper --parallel --full-output <<'EOF'
...
EOF
```

**Output Modes:**
- **Summary (default)**: Structured report with changes, output, verification, and review summary.
- **Full (`--full-output`)**: Complete task messages. Use only when debugging specific failures.

**With per-task backend**:
```bash
fish-agent-wrapper --parallel <<'EOF'
---TASK---
id: task1
backend: codex
workdir: /path/to/dir
---CONTENT---
analyze code structure
---TASK---
id: task2
backend: claude
dependencies: task1
---CONTENT---
design architecture based on analysis
---TASK---
id: task3
backend: gemini
dependencies: task2
---CONTENT---
generate implementation code
EOF
```

If review/retry task is needed, set `backend: ampcode` for that task.

**Concurrency Control**:
Set `FISH_AGENT_WRAPPER_MAX_PARALLEL_WORKERS` to limit concurrent tasks (default: unlimited).

## Environment Variables

- `CODEX_TIMEOUT`: Override timeout in milliseconds (default: 7200000 = 2 hours)
- `FISH_AGENT_WRAPPER_SKIP_PERMISSIONS`: Control Claude CLI permission checks
  - For **Claude** backend: default is **skip permissions** unless explicitly disabled
  - Set `FISH_AGENT_WRAPPER_SKIP_PERMISSIONS=false` to keep Claude permission prompts
- `FISH_AGENT_WRAPPER_MAX_PARALLEL_WORKERS`: Limit concurrent tasks in parallel mode (default: unlimited, recommended: 8)
- `FISH_AGENT_WRAPPER_CLAUDE_DIR`: Override the base Claude config dir (default: `~/.claude`)
- `FISH_AGENT_WRAPPER_AMPCODE_MODE`: Set Ampcode mode (`smart|deep|rush|free`, default: `smart`)

## Invocation Pattern

**Single Task**:
```
Bash tool parameters:
- command: fish-agent-wrapper --backend <backend> - [working_dir] <<'EOF'
  <task content>
  EOF
- timeout: 7200000
- description: <brief description>

Note: `--backend` is recommended; supported values: `codex | claude | gemini | ampcode` (default: `codex`)
```

**Parallel Tasks**:
```
Bash tool parameters:
- command: fish-agent-wrapper --parallel --backend <backend> <<'EOF'
  ---TASK---
  id: task_id
  backend: <backend>  # Optional, overrides global
  workdir: /path
  dependencies: dep1, dep2
  ---CONTENT---
  task content
  EOF
- timeout: 7200000
- description: <brief description>

Note: Global --backend is required; per-task backend is optional
```

## Critical Rules

**NEVER kill fish-agent-wrapper processes.** Long-running tasks are normal. Instead:

1. **Check task status via log file**:
   ```bash
   # View real-time output
   tail -f /tmp/claude/<workdir>/tasks/<task_id>.output

   # Check if task is still running
   cat /tmp/claude/<workdir>/tasks/<task_id>.output | tail -50
   ```

2. **Wait with timeout**:
   ```bash
   # Use TaskOutput tool with block=true and timeout
   TaskOutput(task_id="<id>", block=true, timeout=300000)
   ```

3. **Check process without killing**:
   ```bash
   ps aux | grep fish-agent-wrapper | grep -v grep
   ```

**Why:** fish-agent-wrapper tasks often take 2-10 minutes. Killing them wastes API costs and loses progress.

## Security Best Practices

- **Claude Backend**: Permission checks enabled by default
  - To skip checks: set `FISH_AGENT_WRAPPER_SKIP_PERMISSIONS=true` or pass `--skip-permissions`
- **Concurrency Limits**: Set `FISH_AGENT_WRAPPER_MAX_PARALLEL_WORKERS` in production to prevent resource exhaustion
- **Automation Context**: This wrapper is designed for AI-driven automation where permission prompts would block execution

## Recent Updates

- Multi-backend support for all modes (workdir, resume, parallel)
- Security controls with configurable permission checks
- Concurrency limits with worker pool and fail-fast cancellation
- Ampcode backend support for new/resume/parallel execution
