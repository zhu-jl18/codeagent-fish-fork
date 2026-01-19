#!/usr/bin/env python3
"""Dev-only uninstaller for myclaude-simple (personal setup).

Removes only the files installed by ./install.py and leaves unrelated user files intact.
"""

from __future__ import annotations

import argparse
import os
import sys
from pathlib import Path


DEFAULT_INSTALL_DIR = "~/.claude"
BACKENDS = ("codex", "claude", "gemini", "opencode")

CLAUDE_BLOCK_BEGIN = "<!-- BEGIN MYCLAUDE-SIMPLE:DEV-ONLY -->"
CLAUDE_BLOCK_END = "<!-- END MYCLAUDE-SIMPLE:DEV-ONLY -->"


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Uninstall dev-only workflow + codeagent-wrapper")
    p.add_argument(
        "--install-dir",
        default=DEFAULT_INSTALL_DIR,
        help="Install directory (default: ~/.claude)",
    )
    p.add_argument(
        "-y",
        "--yes",
        action="store_true",
        help="Do not prompt (non-interactive).",
    )
    return p.parse_args(argv)


def _unlink(path: Path) -> bool:
    try:
        path.unlink()
        return True
    except FileNotFoundError:
        return False
    except IsADirectoryError:
        return False


def _rmdir_if_empty(path: Path) -> None:
    try:
        if path.is_dir() and not any(path.iterdir()):
            path.rmdir()
    except OSError:
        return


def _strip_managed_claude_block(text: str) -> tuple[str, bool]:
    start = text.find(CLAUDE_BLOCK_BEGIN)
    if start == -1:
        return text, False
    end = text.find(CLAUDE_BLOCK_END, start)
    if end == -1:
        return text[:start], True
    end += len(CLAUDE_BLOCK_END)
    return text[:start] + text[end:], True


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    install_dir = Path(args.install_dir).expanduser().resolve()

    if not install_dir.exists():
        print(f"Install dir not found: {install_dir}")
        return 0

    if not args.yes:
        print(f"About to remove dev-only installed files from: {install_dir}")
        print("Proceed? [y/N] ", end="", flush=True)
        if input().strip().lower() not in ("y", "yes"):
            print("Aborted.")
            return 0

    exe_name = "codeagent-wrapper.exe" if os.name == "nt" else "codeagent-wrapper"

    targets = [
        install_dir / "commands" / "dev.md",
        install_dir / "agents" / "dev-plan-generator.md",
        install_dir / "skills" / "codeagent" / "SKILL.md",
        install_dir / "skills" / "product-requirements" / "SKILL.md",
        install_dir / "bin" / exe_name,
    ]

    removed = 0

    claude_md = install_dir / "CLAUDE.md"
    if claude_md.exists():
        existing = claude_md.read_text(encoding="utf-8")
        new, changed = _strip_managed_claude_block(existing)
        if changed:
            new_stripped = new.strip()
            if new_stripped == "":
                if _unlink(claude_md):
                    removed += 1
                    print(f"Removed: {claude_md} (managed block; file deleted)")
            else:
                claude_md.write_text(new.rstrip() + "\n", encoding="utf-8")
                print(f"Updated: {claude_md} (removed managed dev-only rules)")

    for path in targets:
        if _unlink(path):
            removed += 1
            print(f"Removed: {path}")

    codeagent_dir = install_dir / "codeagent"
    for backend in BACKENDS:
        _unlink(codeagent_dir / f"{backend}-prompt.md")
    _rmdir_if_empty(codeagent_dir)

    _rmdir_if_empty(install_dir / "skills" / "codeagent")
    _rmdir_if_empty(install_dir / "skills" / "product-requirements")
    _rmdir_if_empty(install_dir / "skills")
    _rmdir_if_empty(install_dir / "commands")
    _rmdir_if_empty(install_dir / "agents")
    _rmdir_if_empty(install_dir / "bin")
    _rmdir_if_empty(install_dir)

    if removed == 0:
        print("Nothing removed (targets not found).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
