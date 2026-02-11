# code-router

<p align="center">
  <a href="README.md">中文</a> | <strong>English</strong>
</p>

Fork notice:
- This is a personal, heavily simplified fork derived from `cexll/myclaude`.
- Scope: `/dev` workflow + `code-router` + PRD skill. Everything else is intentionally removed.

What you get:
- `/dev` workflow (requirements -> plan -> parallel execution -> verification)
- `product-requirements` skill (PRD generator)
- `code-router` (Go executor; backends: `codex` / `claude` / `gemini`; core: `--parallel`)

## Install (WSL2/Linux + macOS + Windows)

```bash
python3 install.py
```

Optional:
```bash
python3 install.py --install-dir ~/.code-router --force
python3 install.py --skip-wrapper
```

Installer outputs:
- `~/.code-router/.env` (single runtime config source)
- `~/.code-router/prompts/*-prompt.md` (per-backend placeholders)
- `~/.code-router/bin/code-router` (or `.exe` on Windows)

Not automated (manual by design):
- No auto-copy of `skills/`, `dev-workflow/commands`, or `dev-workflow/agents` into your target CLI root/project scope
- Manually copy what you need based on your target CLI:
  - **Skills**: pick from `skills/*` (for example: `skills/dev`, `skills/code-router`)
  - **/dev command (Claude Code, etc.)**: use `dev-workflow/commands/dev.md` and `dev-workflow/agents/*`

## Maintain (Rebuild Dist Binaries)

```bash
bash scripts/build-dist.sh
```

This produces:
- `dist/code-router-linux-amd64`
- `dist/code-router-darwin-arm64`
- `dist/code-router-windows-amd64.exe`

## Prompt Injection (Default-On, Empty = No-Op)

Default prompt placeholder files:
- `~/.code-router/prompts/codex-prompt.md`
- `~/.code-router/prompts/claude-prompt.md`
- `~/.code-router/prompts/gemini-prompt.md`

Behavior:
- Wrapper loads the per-backend prompt and prepends it only if it has non-empty content.
- Empty/whitespace-only or missing prompt files behave like "no injection".

Runtime behavior (approval/bypass flags, timeout, parallel propagation rules):
- `docs/runtime-config.md`

## Usage

In Claude Code:
```text
/dev "implement X"
```

PRD:
```text
/product-requirements "write a PRD for feature X"
```

## Dev

```bash
cd code-router
go test ./...
```
