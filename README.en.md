# code-dispatcher

<p align="center">
  <a href="README.md">中文</a> | <strong>English</strong>
</p>

> Receive task → select backend → build args → dispatch execution → collect results. That's dispatch.

Fork notice:
- This is a personal, heavily simplified fork derived from `cexll/myclaude`.
- Scope: `/dev` workflow + `code-dispatcher`. Everything else is intentionally removed.

What you get:
- `/dev` workflow (requirements -> plan -> parallel execution -> verification)
- `code-council` skill (multi-perspective parallel code review with host agent final pass)
- `code-dispatcher` (Go executor; backends: `codex` / `claude` / `gemini`; core: `--parallel`)

## Backend Positioning (Recommended)

- `codex`: default implementation backend (complex logic, multi-file refactors, debugging)
- `claude`: fast fixes, config updates, documentation cleanup
- `gemini`: frontend UI/UX prototyping, styling, and interaction polish
- Invocation rule: all backends must be invoked through `code-dispatcher`; do not call `codex` / `claude` / `gemini` directly.

## Install (WSL2/Linux + macOS + Windows)

Default install path: download the current-platform binary from GitHub Release tag `latest` (no Go required).

```bash
python3 install.py
```

Optional:
```bash
python3 install.py --install-dir ~/.code-dispatcher --force
python3 install.py --skip-dispatcher
python3 install.py --repo zhu-jl18/code-dispatcher --release-tag latest
```

Installer outputs:
- `~/.code-dispatcher/.env` (single runtime config source)
- `~/.code-dispatcher/prompts/*-prompt.md` (per-backend placeholders)
- `~/.code-dispatcher/bin/code-dispatcher` (or `.exe` on Windows)

Not automated (manual by design):
- No auto-copy of `skills/`, `dev-workflow/commands`, or `dev-workflow/agents` into your target CLI root/project scope
- Manually copy what you need based on your target CLI:
  - **Skills**: pick from `skills/*` (for example: `skills/dev`, `skills/code-dispatcher`, `skills/code-council`)
  - **/dev command (Claude Code, etc.)**: use `dev-workflow/commands/dev.md` and `dev-workflow/agents/*`

Notes:
- `install.py` requires network access to GitHub Releases for binary installation.
- Use `--skip-dispatcher` if you only need runtime config/assets.

## Local Build (Optional)

```bash
bash scripts/build-dist.sh
```

Local artifacts (not tracked by git by default):
- `dist/code-dispatcher-linux-amd64`
- `dist/code-dispatcher-darwin-arm64`
- `dist/code-dispatcher-windows-amd64.exe`

## Prompt Injection (Default-On, Empty = No-Op)

Default prompt placeholder files:
- `~/.code-dispatcher/prompts/codex-prompt.md`
- `~/.code-dispatcher/prompts/claude-prompt.md`
- `~/.code-dispatcher/prompts/gemini-prompt.md`

Behavior:
- code-dispatcher loads the per-backend prompt and prepends it only if it has non-empty content.
- Empty/whitespace-only or missing prompt files behave like "no injection".

Runtime behavior (approval/bypass flags, timeout, parallel propagation rules):
- `docs/runtime-config.md`

## Usage

In Claude Code:
```text
/dev "implement X"
```

Code review:
```text
Review @src/auth/ using code-council
```

## Dev

```bash
cd code-dispatcher
go test ./...
```
