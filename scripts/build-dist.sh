#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$ROOT/dist"

mkdir -p "$OUT_DIR"

echo "[build-dist] building code-router artifacts into: $OUT_DIR"

(
  cd "$ROOT/code-router"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o "$OUT_DIR/code-router-linux-amd64"
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o "$OUT_DIR/code-router-windows-amd64.exe"
  CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o "$OUT_DIR/code-router-darwin-arm64"
)

chmod +x "$OUT_DIR/code-router-linux-amd64" "$OUT_DIR/code-router-darwin-arm64" || true

echo "[build-dist] done:"
ls -la "$OUT_DIR" | sed -n '1,200p'
