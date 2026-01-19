#!/usr/bin/env python3
"""Dev-only installer for myclaude-simple (personal setup).

Installs:
- /dev workflow (command + dev-plan-generator agent)
- codeagent skill
- product-requirements (PRD) skill
- Append dev-only hard gates to CLAUDE.md (non-destructive)
- codeagent-wrapper binary (copied from prebuilt artifacts in ./dist)
- per-backend prompt placeholder files (empty by default)

Targets:
- WSL2/Linux
- Windows
"""

from __future__ import annotations

import argparse
import os
import shutil
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent
DEFAULT_INSTALL_DIR = "~/.claude"

BACKENDS = ("codex", "claude", "gemini", "opencode")

CLAUDE_BLOCK_BEGIN = "<!-- BEGIN MYCLAUDE-SIMPLE:DEV-ONLY -->"
CLAUDE_BLOCK_END = "<!-- END MYCLAUDE-SIMPLE:DEV-ONLY -->"


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Install dev-only Claude Code workflow + codeagent-wrapper")
    p.add_argument(
        "--install-dir",
        default=DEFAULT_INSTALL_DIR,
        help="Install directory (default: ~/.claude)",
    )
    p.add_argument(
        "--force",
        action="store_true",
        help="Overwrite existing files and refresh the managed CLAUDE.md block",
    )
    p.add_argument(
        "--skip-wrapper",
        "--skip-build",
        action="store_true",
        help="Skip installing codeagent-wrapper (only install config/assets)",
    )
    return p.parse_args(argv)


def _ensure_dir(path: Path) -> None:
    path.mkdir(parents=True, exist_ok=True)


def _copy_file(src: Path, dst: Path, *, force: bool) -> None:
    _ensure_dir(dst.parent)
    if dst.exists() and not force:
        return
    shutil.copy2(src, dst)


def _write_if_missing(path: Path, content: str, *, force: bool) -> None:
    _ensure_dir(path.parent)
    if path.exists() and not force:
        return
    path.write_text(content, encoding="utf-8")


def _strip_managed_claude_block(text: str) -> str:
    start = text.find(CLAUDE_BLOCK_BEGIN)
    if start == -1:
        return text
    end = text.find(CLAUDE_BLOCK_END, start)
    if end == -1:
        return text[:start]
    end += len(CLAUDE_BLOCK_END)
    return text[:start] + text[end:]


def _apply_claude_md(install_dir: Path, *, force: bool) -> bool:
    add_path = REPO_ROOT / "memory" / "CLAUDE-add.md"
    add = add_path.read_text(encoding="utf-8").rstrip()

    block = f"{CLAUDE_BLOCK_BEGIN}\n{add}\n{CLAUDE_BLOCK_END}\n"

    dst = install_dir / "CLAUDE.md"
    existing = dst.read_text(encoding="utf-8") if dst.exists() else ""
    if CLAUDE_BLOCK_BEGIN in existing and not force:
        return False

    base = _strip_managed_claude_block(existing).rstrip()
    if base:
        base += "\n\n"

    _ensure_dir(dst.parent)
    dst.write_text(base + block, encoding="utf-8")
    return True


def _install_prompts(install_dir: Path, *, force: bool) -> None:
    codeagent_dir = install_dir / "codeagent"
    _ensure_dir(codeagent_dir)
    for backend in BACKENDS:
        _write_if_missing(codeagent_dir / f"{backend}-prompt.md", "", force=force)


def _copy_prebuilt_wrapper(install_dir: Path, *, force: bool) -> Path:
    bin_dir = install_dir / "bin"
    _ensure_dir(bin_dir)
    exe_name = "codeagent-wrapper.exe" if os.name == "nt" else "codeagent-wrapper"
    out = bin_dir / exe_name

    artifact_name = "codeagent-wrapper-windows-amd64.exe" if os.name == "nt" else "codeagent-wrapper-linux-amd64"
    artifact = REPO_ROOT / "dist" / artifact_name
    if not artifact.exists():
        raise FileNotFoundError(f"missing prebuilt artifact: {artifact}")

    if out.exists() and not force:
        return out

    shutil.copy2(artifact, out)
    if os.name != "nt":
        try:
            out.chmod(0o755)
        except OSError:
            pass
    return out


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    install_dir = Path(args.install_dir).expanduser().resolve()

    _ensure_dir(install_dir)

    claude_updated = _apply_claude_md(install_dir, force=args.force)

    _copy_file(
        REPO_ROOT / "dev-workflow" / "commands" / "dev.md",
        install_dir / "commands" / "dev.md",
        force=args.force,
    )
    _copy_file(
        REPO_ROOT / "dev-workflow" / "agents" / "dev-plan-generator.md",
        install_dir / "agents" / "dev-plan-generator.md",
        force=args.force,
    )

    _copy_file(
        REPO_ROOT / "skills" / "codeagent" / "SKILL.md",
        install_dir / "skills" / "codeagent" / "SKILL.md",
        force=args.force,
    )
    _copy_file(
        REPO_ROOT / "skills" / "product-requirements" / "SKILL.md",
        install_dir / "skills" / "product-requirements" / "SKILL.md",
        force=args.force,
    )

    _install_prompts(install_dir, force=args.force)

    wrapper_path: Path | None = None
    if not args.skip_wrapper:
        try:
            wrapper_path = _copy_prebuilt_wrapper(install_dir, force=args.force)
        except FileNotFoundError as e:
            print(f"ERROR: {e}", file=sys.stderr)
            print("Hint: run `bash scripts/build-dist.sh` in the repo root to generate ./dist artifacts.", file=sys.stderr)
            return 1

    print(f"Installed to: {install_dir}")
    if claude_updated:
        print(f"- claude:   {install_dir / 'CLAUDE.md'} (appended managed dev-only rules)")
    else:
        print(f"- claude:   {install_dir / 'CLAUDE.md'} (kept existing managed rules; use --force to refresh)")
    print(f"- commands: {install_dir / 'commands'}")
    print(f"- agents:   {install_dir / 'agents'}")
    print(f"- skills:   {install_dir / 'skills'}")
    print(f"- prompts:  {install_dir / 'codeagent'} (*-prompt.md placeholders)")
    if wrapper_path is not None:
        print(f"- wrapper:  {wrapper_path} (copied from ./dist)")

    if str(install_dir) != str(Path(DEFAULT_INSTALL_DIR).expanduser().resolve()):
        print("")
        print("Note:")
        print(f"- You used a non-default install dir.")
        print(f"- Set CODEAGENT_CLAUDE_DIR={install_dir} so codeagent-wrapper can find prompts/settings.")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
