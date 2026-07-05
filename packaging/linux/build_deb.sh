#!/usr/bin/env bash
# Скрипт сборки .deb пакета для Топор
set -euo pipefail

VERSION="1.0.0"
PACKAGE_NAME="topor"
ARCH="amd64"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/dist/deb-build"
PKG_DIR="$BUILD_DIR/${PACKAGE_NAME}_${VERSION}_${ARCH}"

echo "=== Сборка .deb пакета Топор v${VERSION} ==="

# Очистка
rm -rf "$BUILD_DIR"
mkdir -p "$PKG_DIR/DEBIAN"
mkdir -p "$PKG_DIR/opt/topor"
mkdir -p "$PKG_DIR/usr/bin"
mkdir -p "$PKG_DIR/usr/share/applications"

# Сборка исполняемого файла через PyInstaller
cd "$PROJECT_ROOT"
pip install pyinstaller
pyinstaller --noconfirm --onedir --name topor \
    --hidden-import=topor \
    --hidden-import=topor.core \
    --hidden-import=topor.gui \
    src/topor/__main__.py

# Копирование файлов
cp -r "$PROJECT_ROOT/dist/topor/"* "$PKG_DIR/opt/topor/"
cp "$SCRIPT_DIR/topor.desktop" "$PKG_DIR/usr/share/applications/"

# Симлинк
ln -sf /opt/topor/topor "$PKG_DIR/usr/bin/topor"

# Управляющий файл
cat > "$PKG_DIR/DEBIAN/control" <<EOF
Package: ${PACKAGE_NAME}
Version: ${VERSION}
Architecture: ${ARCH}
Maintainer: Горшков Сергей Владимирович <nookbat@gmail.com>
Description: Топор — массовый редактор метаданных и временных штампов файлов
 Кроссплатформенное десктопное приложение для пакетного изменения
 атрибутов файлов офисных и текстовых форматов.
Homepage: https://nookbat.ru/
Section: utils
Priority: optional
Depends: libc6 (>= 2.31), libgl1, libxcb-xinerama0, libxkbcommon0
EOF

# Скрипты
cat > "$PKG_DIR/DEBIAN/postinst" <<'EOF'
#!/bin/sh
set -e
update-desktop-database /usr/share/applications 2>/dev/null || true
EOF
chmod 755 "$PKG_DIR/DEBIAN/postinst"

# Сборка пакета
dpkg-deb --build "$PKG_DIR"
mv "$BUILD_DIR/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb" "$PROJECT_ROOT/dist/"

echo "=== Готово: dist/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb ==="
