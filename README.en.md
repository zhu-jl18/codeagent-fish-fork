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

Notes:
- The installer copies a prebuilt `fish-agent-wrapper` binary from `./dist` (no Go toolchain required at install time).
- It appends a managed workflow block to your `CLAUDE.md` (non-destructive; `--force` refreshes the managed block).

Optional:
```bash
python3 install.py --install-dir ~/.claude --force
python3 install.py --skip-wrapper
```

It installs/updates:
- `CLAUDE.md` (append-only managed block)
- `commands/dev.md`
- `agents/dev-plan-generator.md`
- `skills/fish-agent-wrapper/SKILL.md`
- `skills/product-requirements/SKILL.md`
- `~/.claude/fish-agent-wrapper/*-prompt.md` (per-backend empty placeholders; used for prompt injection)
- `~/.claude/bin/fish-agent-wrapper` (or `.exe` on Windows)

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
- `~/.claude/fish-agent-wrapper/codex-prompt.md`
- `~/.claude/fish-agent-wrapper/claude-prompt.md`
- `~/.claude/fish-agent-wrapper/gemini-prompt.md`

Behavior:
- Wrapper loads the per-backend prompt and prepends it only if it has non-empty content.
- Empty/whitespace-only or missing prompt files behave like "no injection".

Useful env vars:
- `FISH_AGENT_WRAPPER_CLAUDE_DIR`: base dir (default `~/.claude`)

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
