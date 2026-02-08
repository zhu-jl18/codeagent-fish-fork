# fish-agent-wrapper Runtime Config

This document is the single source of truth for runtime behavior related to:

- approval / bypass flags per backend
- environment variables
- timeout behavior
- parallel-mode propagation rules

## 1) Backend Approval/Bypass Matrix

| Backend | Wrapper-added flag | Default behavior | Disable/Change path |
|---|---|---|---|
| `codex` | `--dangerously-bypass-approvals-and-sandbox` | enabled by default | set `CODEX_BYPASS_SANDBOX=false` |
| `claude` | `--dangerously-skip-permissions` | enabled by default | set `FISH_AGENT_WRAPPER_SKIP_PERMISSIONS=false` |
| `gemini` | `-y` (`--yolo`) | always enabled in current wrapper | no wrapper env toggle currently |

Notes:

- "enabled by default" means env is unset, or explicitly true-like.
- true-like values: `1`, `true`, `yes`, `on`.
- false-like values: `0`, `false`, `no`, `off`.

## 2) Environment Variables

### Approval/Bypass Controls

- `CODEX_BYPASS_SANDBOX` (codex only)
  - unset/true => wrapper appends codex dangerous bypass flag
  - false => wrapper does not append codex dangerous bypass flag

- `FISH_AGENT_WRAPPER_SKIP_PERMISSIONS` (claude)
  - unset/true => wrapper appends:
    - claude: `--dangerously-skip-permissions`
  - false => wrapper keeps permission prompts for claude

### Runtime Controls

- `CODEX_TIMEOUT`
  - default: `7200000` (2 hours)
  - parser behavior:
    - value `>10000` => treated as milliseconds, converted to seconds
    - value `<=10000` => treated as seconds

- `FISH_AGENT_WRAPPER_MAX_PARALLEL_WORKERS`
  - default: unlimited (`0`)
  - recommended: `8`
  - hard cap in wrapper: `100`

- `FISH_AGENT_WRAPPER_CLAUDE_DIR`
  - default: `~/.claude`
  - used for prompt file + Claude settings resolution

## 3) Parallel Propagation Rule (`skip_permissions`)

Parallel mode computes a global skip-permissions value from CLI/env, then applies:

```text
effective_task_skip = task.skip_permissions OR global_skip_permissions
```

Implication:

- If global skip-permissions is true, a task cannot force it back to false.

## 4) Timeout Layering (Important)

There are usually two timeout layers:

- outer caller timeout (e.g., tool invocation timeout)
- wrapper timeout (`CODEX_TIMEOUT`)

Effective timeout is whichever triggers first.

If wrapper shows 2h but execution still stops around 5m, the outer caller timeout is likely lower.

## 5) How to Set Variables

### One-off (single command)

```bash
KEY=value fish-agent-wrapper --backend codex "task"
```

### Current shell session

```bash
export KEY=value
```

### Persistent (bash)

Add to `~/.bashrc`:

```bash
export KEY=value
```

### Persistent (fish)

```fish
set -Ux KEY value
```

## 6) Practical Presets

### Default full-bypass behavior (project-preferred)

- Do not set any of these toggles:
  - `CODEX_BYPASS_SANDBOX`
  - `FISH_AGENT_WRAPPER_SKIP_PERMISSIONS`

### Stricter approval behavior

```bash
export CODEX_BYPASS_SANDBOX=false
export FISH_AGENT_WRAPPER_SKIP_PERMISSIONS=false
```

Note: gemini still runs with `-y` in current wrapper implementation.
