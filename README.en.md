# fish-agent-wrapper

<p align="center">
  <a href="README.md">中文</a> | <strong>English</strong>
</p>

Fork notice:
- This is a personal, heavily simplified fork derived from `cexll/myclaude`.
- Scope: `/dev` workflow + `fish-agent-wrapper` + PRD skill. Everything else is intentionally removed.

What you get:
- `/dev` workflow (requirements -> plan -> parallel execution -> verification)
- `product-requirements` skill (PRD generator)
- `fish-agent-wrapper` (Go executor; backends: `codex` / `claude` / `gemini`; core: `--parallel`)

## Install (WSL2/Linux + macOS + Windows)

```bash
python3 install.py
```

Optional:
```bash
python3 install.py --install-dir ~/.fish-agent-wrapper --force
python3 install.py --skip-wrapper
```

Installer outputs:
- `~/.fish-agent-wrapper/.env` (single runtime config source)
- `~/.fish-agent-wrapper/prompts/*-prompt.md` (per-backend placeholders)
- `~/.fish-agent-wrapper/bin/fish-agent-wrapper` (or `.exe` on Windows)

Not automated (manual by design):
- No auto-copy of `skills/`, `dev-workflow/commands`, or `dev-workflow/agents` into your target CLI root/project scope
- Manually copy what you need based on your target CLI:
  - **Skills**: pick from `skills/*` (for example: `skills/dev`, `skills/fish-agent-wrapper`)
  - **/dev command (Claude Code, etc.)**: use `dev-workflow/commands/dev.md` and `dev-workflow/agents/*`

## Maintain (Rebuild Dist Binaries)

```bash
bash scripts/build-dist.sh
```

This produces:
- `dist/fish-agent-wrapper-linux-amd64`
- `dist/fish-agent-wrapper-darwin-arm64`
- `dist/fish-agent-wrapper-windows-amd64.exe`

## Prompt Injection (Default-On, Empty = No-Op)

Default prompt placeholder files:
- `~/.fish-agent-wrapper/prompts/codex-prompt.md`
- `~/.fish-agent-wrapper/prompts/claude-prompt.md`
- `~/.fish-agent-wrapper/prompts/gemini-prompt.md`

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
cd fish-agent-wrapper
go test ./...
```
