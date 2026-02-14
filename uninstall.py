#!/usr/bin/env python3
"""Uninstaller for code-dispatcher runtime assets.

Removes only the files installed by ./install.py and leaves unrelated user files intact.
"""

from __future__ import annotations

import argparse
import os
from pathlib import Path


DEFAULT_INSTALL_DIR = "~/.code-dispatcher"
BACKENDS = ("codex", "claude", "gemini")
LEGACY_PROMPT_FILES = ("copilot-prompt.md",)


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Uninstall code-dispatcher")
    p.add_argument(
        "--install-dir",
        default=DEFAULT_INSTALL_DIR,
        help="Install directory (default: ~/.code-dispatcher)",
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


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    install_dir = Path(args.install_dir).expanduser().resolve()

    if not install_dir.exists():
        print(f"Install dir not found: {install_dir}")
        return 0

    if not args.yes:
        print(f"About to remove code-dispatcher installed files from: {install_dir}")
        print("Proceed? [y/N] ", end="", flush=True)
        if input().strip().lower() not in ("y", "yes"):
            print("Aborted.")
            return 0

    exe_name = "code-dispatcher.exe" if os.name == "nt" else "code-dispatcher"

    targets = [
        install_dir / ".env",
        install_dir / "bin" / exe_name,
    ]

    removed = 0

    for path in targets:
        if _unlink(path):
            removed += 1
            print(f"Removed: {path}")

    wrapper_dir = install_dir / "prompts"
    prompt_files = [f"{backend}-prompt.md" for backend in BACKENDS]
    prompt_files.extend(LEGACY_PROMPT_FILES)
    for prompt_file in prompt_files:
        path = wrapper_dir / prompt_file
        if _unlink(path):
            removed += 1
            print(f"Removed: {path}")
    _rmdir_if_empty(wrapper_dir)

    _rmdir_if_empty(install_dir / "bin")
    _rmdir_if_empty(install_dir)

    if removed == 0:
        print("Nothing removed (targets not found).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
