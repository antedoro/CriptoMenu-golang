#!/bin/bash

# Configuration
APP_NAME="CriptoMenu"
ICON_PNG="icon.png" # Path to your source PNG icon
BUILD_DIR="./build" # Directory for build artifacts

# Function to build the app for a specific architecture
build_app() {
    local ARCH=$1
    local SUFFIX=$2
    local APP_BUNDLE_NAME="${APP_NAME}_${SUFFIX}.app"
    local FULL_APP_PATH="$BUILD_DIR/$APP_BUNDLE_NAME"
    local MACOS_DIR="$FULL_APP_PATH/Contents/MacOS"
    local RESOURCES_DIR="$FULL_APP_PATH/Contents/Resources"
    local EXECUTABLE_PATH="$MACOS_DIR/$APP_NAME"

    echo "--- Building for $SUFFIX ($ARCH) ---"

    # Create .app bundle structure
    mkdir -p "$MACOS_DIR"
    mkdir -p "$RESOURCES_DIR"

    # Build Go executable
    echo "Compiling Go binary..."
    CGO_ENABLED=1 GOOS=darwin GOARCH=$ARCH go build -o "$EXECUTABLE_PATH" .
    if [ $? -ne 0 ]; then
        echo "Error: Go build failed for $ARCH."
        exit 1
    fi

    # Copy Icon (assuming it was generated previously in BUILD_DIR)
    if [ -f "$BUILD_DIR/AppIcon.icns" ]; then
        cp "$BUILD_DIR/AppIcon.icns" "$RESOURCES_DIR/AppIcon.icns"
    else
        echo "Warning: AppIcon.icns not found."
    fi

    # Create Info.plist
    echo "Creating Info.plist..."
    local APP_NAME_LOWER=$(echo "$APP_NAME" | tr '[:upper:]' '[:lower:]')
    cat << EOF > "$FULL_APP_PATH/Contents/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleExecutable</key>
	<string>$APP_NAME</string>
	<key>CFBundleIconFile</key>
	<string>AppIcon</string>
	<key>CFBundleIdentifier</key>
	<string>com.antedoro.$APP_NAME_LOWER</string>
	<key>CFBundleName</key>
	<string>$APP_NAME</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>1.24.4</string>
	<key>LSUIElement</key>
	<true/>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>
EOF

    # Create PkgInfo
    echo "APPL????" > "$FULL_APP_PATH/Contents/PkgInfo"

    # Touch the app bundle
    touch "$FULL_APP_PATH"

    # Ad-hoc code signing
    echo "Signing app bundle..."
    codesign --force --deep --sign - "$FULL_APP_PATH"
    if [ $? -ne 0 ]; then
        echo "Warning: Code signing failed."
    else
        echo "✔ Signed $APP_BUNDLE_NAME"
    fi

    echo "✔ Created $APP_BUNDLE_NAME"
}

# --- Script Start ---

echo "Starting macOS app build for $APP_NAME..."

# Ensure build directory exists and clean it
mkdir -p "$BUILD_DIR"
rm -rf "$BUILD_DIR"/*

# --- Generate Icon (Once) ---
echo "Generating .icns icon from $ICON_PNG..."
ICONSET_DIR="$BUILD_DIR/$APP_NAME.iconset"
mkdir -p "$ICONSET_DIR"

if [ -f "$ICON_PNG" ]; then
    sips -z 16 16     "$ICON_PNG" --out "$ICONSET_DIR/icon_16x16.png" > /dev/null
    sips -z 32 32     "$ICON_PNG" --out "$ICONSET_DIR/icon_16x16@2x.png" > /dev/null
    sips -z 32 32     "$ICON_PNG" --out "$ICONSET_DIR/icon_32x32.png" > /dev/null
    sips -z 64 64     "$ICON_PNG" --out "$ICONSET_DIR/icon_32x32@2x.png" > /dev/null
    sips -z 128 128   "$ICON_PNG" --out "$ICONSET_DIR/icon_128x128.png" > /dev/null
    sips -z 256 256   "$ICON_PNG" --out "$ICONSET_DIR/icon_128x128@2x.png" > /dev/null
    sips -z 256 256   "$ICON_PNG" --out "$ICONSET_DIR/icon_256x256.png" > /dev/null
    sips -z 512 512   "$ICON_PNG" --out "$ICONSET_DIR/icon_256x256@2x.png" > /dev/null
    sips -z 512 512   "$ICON_PNG" --out "$ICONSET_DIR/icon_512x512.png" > /dev/null
    sips -z 1024 1024 "$ICON_PNG" --out "$ICONSET_DIR/icon_512x512@2x.png" > /dev/null

    iconutil -c icns "$ICONSET_DIR" -o "$BUILD_DIR/AppIcon.icns"
    rm -rf "$ICONSET_DIR"
    echo "Icon generated at $BUILD_DIR/AppIcon.icns"
else
    echo "Error: Icon file not found at $ICON_PNG."
    exit 1
fi

# --- Build Versions ---

# 1. Build for Intel (amd64)
build_app "amd64" "Intel"

# 2. Build for Apple Silicon (arm64)
build_app "arm64" "AppleSilicon"

# Clean up common icon
rm -f "$BUILD_DIR/AppIcon.icns"

echo "=========================================="
echo "Build complete!"
echo "Artifacts:"
echo " - $BUILD_DIR/${APP_NAME}_Intel.app"
echo " - $BUILD_DIR/${APP_NAME}_AppleSilicon.app"
echo "=========================================="