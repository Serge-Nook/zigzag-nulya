#!/usr/bin/env bash
# Builds a Debian package from an already compiled Linux binary.
set -euo pipefail

VERSION="${VERSION:-0.1.0}"
BIN="${BIN:-dist/topor-linux-amd64}"
OUT="${OUT:-dist}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT

install -Dm755 "$ROOT/$BIN" "$STAGE/usr/bin/topor"
install -Dm644 "$ROOT/packaging/topor.desktop" "$STAGE/usr/share/applications/topor.desktop"
install -Dm644 "$ROOT/packaging/topor.png" "$STAGE/usr/share/icons/hicolor/512x512/apps/topor.png"

mkdir -p "$STAGE/DEBIAN"
cat > "$STAGE/DEBIAN/control" <<EOF
Package: topor
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Depends: libc6, libgl1, libx11-6, libxrandr2, libxcursor1, libxinerama1, libxi6, libxxf86vm1, zenity
Maintainer: Serge Nook <nookbat.ru>
Description: T0P0R
 Изменение дат создания и изменения файлов, а также данных об авторе
 и последнем редакторе документов Word и Excel.
EOF

mkdir -p "$ROOT/$OUT"
dpkg-deb --build --root-owner-group "$STAGE" "$ROOT/$OUT/topor_${VERSION}_amd64.deb"
