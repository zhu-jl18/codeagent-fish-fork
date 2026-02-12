# code-router Runtime Config

This document is the single source of truth for runtime behavior related to:

- approval / bypass flags per backend
- timeout behavior
- parallel-mode propagation rules
- config loading model

## 1) Single Config Source

All runtime options are loaded from:

```text
~/.code-router/.env
```

The wrapper does not read these control options from shell environment variables anymore.

## 2) Backend Approval/Bypass Matrix

| Backend | Wrapper-added flag | Default behavior | Disable/Change path |
|---|---|---|---|
| `codex` | `--dangerously-bypass-approvals-and-sandbox` | enabled by default | set `CODEX_BYPASS_SANDBOX=false` in `.env` |
| `claude` | `--dangerously-skip-permissions` | enabled by default | set `CODE_ROUTER_SKIP_PERMISSIONS=false` in `.env` |
| `copilot` | `--allow-all` | enabled by default | set `CODE_ROUTER_SKIP_PERMISSIONS=false` in `.env` |
| `gemini` | `-y` (`--yolo`) | always enabled in current wrapper | no wrapper toggle currently |

Notes:

- true-like values: `1`, `true`, `yes`, `on`.
- false-like values: `0`, `false`, `no`, `off`.

## 3) Runtime Keys in `.env`

### Approval/Bypass Controls

- `CODEX_BYPASS_SANDBOX` (codex only)
  - unset/true => wrapper appends codex dangerous bypass flag
  - false => wrapper does not append codex dangerous bypass flag

- `CODE_ROUTER_SKIP_PERMISSIONS` (claude/copilot)
  - unset/true => claude appends `--dangerously-skip-permissions`, copilot appends `--allow-all`
  - false => claude keeps permission prompts, copilot drops `--allow-all`

### Runtime Controls

- `CODEX_TIMEOUT`
  - default: `7200000` (2 hours)
  - parser behavior:
    - value `>10000` => treated as milliseconds, converted to seconds
    - value `<=10000` => treated as seconds

- `CODE_ROUTER_MAX_PARALLEL_WORKERS`
  - default: unlimited (`0`)
  - recommended: `8`
  - hard cap in wrapper: `100`

- `CODE_ROUTER_ASCII_MODE`
  - `true` => ASCII status (`PASS/WARN/FAIL`)
  - otherwise => Unicode status symbols

- `CODE_ROUTER_LOGGER_CLOSE_TIMEOUT_MS`
  - default: `5000`
  - `0` => wait indefinitely

### Prompt Files

Prompt files are resolved from:

```text
~/.code-router/prompts/<backend>-prompt.md
```

Supported backends: `codex`, `claude`, `gemini`, `copilot`.

## 4) Parallel Propagation Rule (`skip_permissions`)

Parallel mode computes a global skip-permissions value from CLI/.env, then applies:

```text
effective_task_skip = task.skip_permissions OR global_skip_permissions
```

Implication:

- If global skip-permissions is true, a task cannot force it back to false.

## 5) Timeout Layering (Important)

There are usually two timeout layers:

- outer caller timeout (e.g., tool invocation timeout)
- wrapper timeout (`CODEX_TIMEOUT` from `.env`)

Effective timeout is whichever triggers first.

## 6) Editing Config

Edit the file directly:

```bash
${EDITOR:-vi} ~/.code-router/.env
```
