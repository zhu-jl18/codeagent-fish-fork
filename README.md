# codeagent-fish-fork (dev-only)

Fork notice:
- This is a personal, heavily simplified fork derived from `cexll/myclaude`.
- Scope: dev-only workflow + codeagent-wrapper + PRD skill. Everything else is intentionally removed.
- 中文说明：这是从 `cexll/myclaude` fork 并大幅裁剪的个人版本，仅保留 dev 工作流 + codeagent-wrapper + PRD（product-requirements）skill。

Personal, minimal Claude Code setup:
- `/dev` workflow (requirements -> analysis -> dev plan -> parallel execution -> coverage gate)
- `product-requirements` skill (PRD generator)
- `codeagent-wrapper` (Go executor; backends: codex/claude/gemini/opencode; core: `--parallel`)

Everything else from the original upstream repo is intentionally removed.

## Install (WSL2/Linux + Windows)

```bash
python3 install.py
```

This installer copies a prebuilt `codeagent-wrapper` binary from `./dist` (no Go toolchain required at install time).

Optional:
```bash
python3 install.py --install-dir ~/.claude --force
python3 install.py --skip-wrapper
```

The installer installs/updates:
- `CLAUDE.md` (appends managed dev-only rules; non-destructive, `--force` refreshes the block)
- `commands/dev.md`
- `agents/dev-plan-generator.md`
- `skills/codeagent/SKILL.md`
- `skills/product-requirements/SKILL.md`

And builds:
- `~/.claude/bin/codeagent-wrapper` (or `.exe` on Windows)

## Maintain (Update Prebuilt Binaries)

```bash
bash scripts/build-dist.sh
```

## Prompt Injection (Default-On, Empty = No-Op)

Default prompt placeholder files:
- `~/.claude/codeagent/codex-prompt.md`
- `~/.claude/codeagent/claude-prompt.md`
- `~/.claude/codeagent/gemini-prompt.md`
- `~/.claude/codeagent/opencode-prompt.md`

Behavior:
- Wrapper loads the per-backend prompt and prepends it **only if it has non-empty content**.
- Empty/whitespace-only or missing prompt files behave like "no injection".

Useful env vars:
- `CODEAGENT_CLAUDE_DIR`: base dir (default `~/.claude`)

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
go test ./...
```

Run inside `codeagent-wrapper/`.
