#!/usr/bin/env python3
"""Installer for code-router runtime assets.

Installs:
- code-router binary (copied from prebuilt artifacts in ./dist)
- per-backend prompt placeholders under ~/.code-router/prompts
- ~/.code-router/.env template for runtime configuration

Targets:
- WSL2/Linux
- macOS (Apple Silicon)
- Windows
"""

from __future__ import annotations

import argparse
import os
import platform
import shutil
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent
DEFAULT_INSTALL_DIR = "~/.code-router"

BACKENDS = ("codex", "claude", "gemini")


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Installer for code-router")
    p.add_argument(
        "--install-dir",
        default=DEFAULT_INSTALL_DIR,
        help="Install directory (default: ~/.code-router)",
    )
    p.add_argument(
        "--force",
        action="store_true",
        help="Overwrite existing files",
    )
    p.add_argument(
        "--skip-wrapper",
        "--skip-build",
        action="store_true",
        help="Skip installing code-router binary (only install runtime config/assets)",
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


def _install_prompts(install_dir: Path, *, force: bool) -> None:
    wrapper_dir = install_dir / "prompts"
    _ensure_dir(wrapper_dir)
    for backend in BACKENDS:
        _write_if_missing(wrapper_dir / f"{backend}-prompt.md", "", force=force)


def _install_env_template(install_dir: Path, *, force: bool) -> None:
    env_template = (
        "# code-router runtime config\n"
        "# Values are loaded only from this file.\n\n"
        "# CODEX_TIMEOUT in milliseconds (default: 7200000)\n"
        "CODEX_TIMEOUT=7200000\n\n"
        "# true/false controls\n"
        "CODEX_BYPASS_SANDBOX=true\n"
        "CODE_ROUTER_SKIP_PERMISSIONS=true\n"
        "CODE_ROUTER_ASCII_MODE=false\n"
        "CODE_ROUTER_MAX_PARALLEL_WORKERS=0\n"
        "CODE_ROUTER_LOGGER_CLOSE_TIMEOUT_MS=5000\n\n"
        "# backend credentials (examples)\n"
        "# ANTHROPIC_API_KEY=\n"
        "# GEMINI_API_KEY=\n"
    )
    _write_if_missing(install_dir / ".env", env_template, force=force)


def _get_artifact_name() -> str:
    """Get the correct artifact name for the current platform."""
    system = platform.system()
    if system == "Windows":
        return "code-router-windows-amd64.exe"
    elif system == "Darwin":
        return "code-router-darwin-arm64"
    else:
        return "code-router-linux-amd64"


def _copy_prebuilt_wrapper(install_dir: Path, *, force: bool) -> Path:
    bin_dir = install_dir / "bin"
    _ensure_dir(bin_dir)
    exe_name = "code-router.exe" if os.name == "nt" else "code-router"
    out = bin_dir / exe_name

    artifact_name = _get_artifact_name()
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


def _get_shell_config_path() -> str | None:
    """Detect shell type and return config file path."""
    shell = os.environ.get("SHELL", "")
    home = Path.home()

    if "zsh" in shell:
        return str(home / ".zshrc")
    elif "bash" in shell:
        # macOS uses .bash_profile, Linux uses .bashrc
        if sys.platform == "darwin":
            return str(home / ".bash_profile")
        return str(home / ".bashrc")
    elif "fish" in shell:
        return str(home / ".config" / "fish" / "config.fish")
    return None


def _print_path_hint(bin_path: Path) -> None:
    """Print PATH setup instructions based on platform and shell."""
    print("")
    print("PATH setup:")

    if os.name == "nt":
        # Windows
        print("  Add to PATH manually:")
        print("  1. Open System Properties > Environment Variables")
        print("  2. Edit 'Path' under User variables")
        print(f"  3. Add: {bin_path}")
        print("")
        print("  Or run in PowerShell (current user):")
        print(f'  [Environment]::SetEnvironmentVariable("Path", $env:Path + ";{bin_path}", "User")')
    else:
        # Linux / macOS
        shell_config = _get_shell_config_path()
        shell = os.environ.get("SHELL", "").split("/")[-1] or "sh"

        if "fish" in shell:
            export_cmd = f'set -gx PATH $PATH "{bin_path}"'
        else:
            export_cmd = f'export PATH="$PATH:{bin_path}"'

        print(f"  {export_cmd}")
        print("")
        if shell_config:
            print(f"  To persist, add to {shell_config}:")
            print(f"  echo '{export_cmd}' >> {shell_config}")
        else:
            print("  Add the export command to your shell config file to persist.")


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    install_dir = Path(args.install_dir).expanduser().resolve()

    _ensure_dir(install_dir)

    _install_prompts(install_dir, force=args.force)
    _install_env_template(install_dir, force=args.force)

    wrapper_path: Path | None = None
    if not args.skip_wrapper:
        try:
            wrapper_path = _copy_prebuilt_wrapper(install_dir, force=args.force)
        except FileNotFoundError as e:
            print(f"ERROR: {e}", file=sys.stderr)
            print("Hint: run `bash scripts/build-dist.sh` in the repo root to generate ./dist artifacts.", file=sys.stderr)
            return 1

    print(f"Installed to: {install_dir}")
    print(f"- env:      {install_dir / '.env'}")
    print(f"- prompts:  {install_dir / 'prompts'} (*-prompt.md placeholders)")
    if wrapper_path is not None:
        print(f"- wrapper:  {wrapper_path} (copied from ./dist)")

    print("")
    print("Manual setup required:")
    print("- Copy dev/skill markdown files manually into each target CLI root or project scope as needed.")
    print("- This installer does not auto-copy skills/commands/agents into Claude/Codex/iFlow/Amp roots.")

    if wrapper_path is not None:
        _print_path_hint(install_dir / "bin")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
