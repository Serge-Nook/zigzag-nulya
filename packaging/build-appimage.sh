#!/usr/bin/env bash
# Builds an AppImage from an already compiled Linux binary.
set -euo pipefail

VERSION="${VERSION:-0.1.0}"
BIN="${BIN:-dist/topor-linux-amd64}"
OUT="${OUT:-dist}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APPDIR="$(mktemp -d)/T0P0R.AppDir"
trap 'rm -rf "$(dirname "$APPDIR")"' EXIT

install -Dm755 "$ROOT/$BIN" "$APPDIR/usr/bin/topor"
install -Dm644 "$ROOT/packaging/topor.desktop" "$APPDIR/usr/share/applications/topor.desktop"
install -Dm644 "$ROOT/packaging/topor.png" "$APPDIR/usr/share/icons/hicolor/512x512/apps/topor.png"
cp "$ROOT/packaging/topor.desktop" "$APPDIR/topor.desktop"
cp "$ROOT/packaging/topor.png" "$APPDIR/topor.png"

cat > "$APPDIR/AppRun" <<'EOF'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
exec "$HERE/usr/bin/topor" "$@"
EOF
chmod +x "$APPDIR/AppRun"

APPIMAGETOOL="${APPIMAGETOOL:-}"
if [ -z "$APPIMAGETOOL" ]; then
  APPIMAGETOOL="$(mktemp -d)/appimagetool"
  curl -sSLo "$APPIMAGETOOL" \
    https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage
  chmod +x "$APPIMAGETOOL"
fi

mkdir -p "$ROOT/$OUT"
ARCH=x86_64 "$APPIMAGETOOL" --appimage-extract-and-run "$APPDIR" "$ROOT/$OUT/T0P0R-${VERSION}-x86_64.AppImage"
