#!/usr/bin/env bash
# Скрипт сборки AppImage для Топор
set -euo pipefail

VERSION="1.0.0"
PACKAGE_NAME="Topor"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/dist/appimage-build"
APP_DIR="$BUILD_DIR/${PACKAGE_NAME}.AppDir"

echo "=== Сборка AppImage Топор v${VERSION} ==="

# Очистка
rm -rf "$BUILD_DIR"
mkdir -p "$APP_DIR/usr/bin"
mkdir -p "$APP_DIR/usr/share/applications"
mkdir -p "$APP_DIR/usr/share/icons/hicolor/256x256/apps"

# Сборка через PyInstaller
cd "$PROJECT_ROOT"
pip install pyinstaller
pyinstaller --noconfirm --onedir --name topor \
    --hidden-import=topor \
    --hidden-import=topor.core \
    --hidden-import=topor.gui \
    src/topor/__main__.py

# Копирование файлов
cp -r "$PROJECT_ROOT/dist/topor/"* "$APP_DIR/usr/bin/"
cp "$SCRIPT_DIR/topor.desktop" "$APP_DIR/"
cp "$SCRIPT_DIR/topor.desktop" "$APP_DIR/usr/share/applications/"

# AppRun
cat > "$APP_DIR/AppRun" <<'EOF'
#!/bin/bash
SELF="$(readlink -f "$0")"
HERE="${SELF%/*}"
exec "${HERE}/usr/bin/topor" "$@"
EOF
chmod +x "$APP_DIR/AppRun"

# Скачивание appimagetool
if [ ! -f "$BUILD_DIR/appimagetool" ]; then
    wget -q "https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage" \
        -O "$BUILD_DIR/appimagetool"
    chmod +x "$BUILD_DIR/appimagetool"
fi

# Сборка AppImage
cd "$BUILD_DIR"
ARCH=x86_64 ./appimagetool "$APP_DIR" "$PROJECT_ROOT/dist/${PACKAGE_NAME}-${VERSION}-x86_64.AppImage"

echo "=== Готово: dist/${PACKAGE_NAME}-${VERSION}-x86_64.AppImage ==="
