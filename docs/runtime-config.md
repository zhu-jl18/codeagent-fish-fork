# code-router Runtime Config

This document is the single source of truth for runtime behavior related to:

- timeout behavior
- parallel-mode propagation rules
- config loading model

## 1) Single Config Source

All runtime options are loaded from:

```text
~/.code-router/.env
```

The wrapper does not read these control options from shell environment variables anymore.

## 2) Backend Approval/Bypass

All backends run with approval bypass hardcoded (no toggle):

- `codex`: `--dangerously-bypass-approvals-and-sandbox`
- `claude`: `--dangerously-skip-permissions`
- `gemini`: `-y`

## 3) Runtime Keys in `.env`

- `CODE_ROUTER_TIMEOUT`
  - default: `7200` (seconds, 2 hours)
  - unit: seconds

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

Supported backends: `codex`, `claude`, `gemini`.

## 4) Timeout Layering (Important)

There are usually two timeout layers:

- outer caller timeout (e.g., tool invocation timeout)
- wrapper timeout (`CODE_ROUTER_TIMEOUT` from `.env`)

Effective timeout is whichever triggers first.

## 5) Editing Config

Edit the file directly:

```bash
${EDITOR:-vi} ~/.code-router/.env
```
