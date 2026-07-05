#!/usr/bin/env bash
# Скрипт сборки .dmg для Топор (macOS)
set -euo pipefail

VERSION="1.0.0"
APP_NAME="Топор"
BUNDLE_NAME="Topor"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/dist/macos-build"
APP_DIR="$BUILD_DIR/${BUNDLE_NAME}.app"

echo "=== Сборка .dmg ${APP_NAME} v${VERSION} ==="

# Очистка
rm -rf "$BUILD_DIR"
mkdir -p "$APP_DIR/Contents/MacOS"
mkdir -p "$APP_DIR/Contents/Resources"

# Сборка через PyInstaller
cd "$PROJECT_ROOT"
pip install pyinstaller
pyinstaller --noconfirm --onedir --name topor \
    --hidden-import=topor \
    --hidden-import=topor.core \
    --hidden-import=topor.gui \
    src/topor/__main__.py

# Копирование файлов
cp -r "$PROJECT_ROOT/dist/topor/"* "$APP_DIR/Contents/MacOS/"

# Info.plist
cat > "$APP_DIR/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleDisplayName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleIdentifier</key>
    <string>ru.nookbat.topor</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundleExecutable</key>
    <string>topor</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>CFBundleDocumentTypes</key>
    <array>
        <dict>
            <key>CFBundleTypeExtensions</key>
            <array>
                <string>docx</string>
                <string>xlsx</string>
                <string>doc</string>
                <string>xls</string>
                <string>odt</string>
                <string>html</string>
                <string>rtf</string>
                <string>txt</string>
                <string>xml</string>
            </array>
            <key>CFBundleTypeName</key>
            <string>Document</string>
            <key>CFBundleTypeRole</key>
            <string>Editor</string>
        </dict>
    </array>
</dict>
</plist>
EOF

# Сборка DMG
hdiutil create -volname "${BUNDLE_NAME}" \
    -srcfolder "$APP_DIR" \
    -ov -format UDZO \
    "$PROJECT_ROOT/dist/${BUNDLE_NAME}-${VERSION}.dmg"

echo "=== Готово: dist/${BUNDLE_NAME}-${VERSION}.dmg ==="
