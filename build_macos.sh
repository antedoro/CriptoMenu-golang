#!/bin/bash

# Configuration
APP_NAME="CriptoMenu"
ICON_PNG="icon.png" # Path to your source PNG icon
BUILD_DIR="./build" # Directory for build artifacts

# --- Script Start ---

echo "Starting macOS app build for $APP_NAME..."

# Ensure build directory exists
mkdir -p "$BUILD_DIR"

# Define output .app bundle path
APP_BUNDLE="$BUILD_DIR/$APP_NAME.app"
MACOS_DIR="$APP_BUNDLE/Contents/MacOS"
RESOURCES_DIR="$APP_BUNDLE/Contents/Resources"
EXECUTABLE_PATH="$MACOS_DIR/$APP_NAME"

# Clean up previous build
echo "Cleaning up previous build artifacts..."
rm -rf "$APP_BUNDLE"
rm -rf "$BUILD_DIR/$APP_NAME.iconset"
rm -f "$BUILD_DIR/$APP_NAME.icns"

# Create .app bundle structure
echo "Creating .app bundle structure..."
mkdir -p "$MACOS_DIR"
mkdir -p "$RESOURCES_DIR"

# Build Go executable
echo "Building Go executable for macOS..."
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o "$EXECUTABLE_PATH" .
if [ $? -ne 0 ]; then
    echo "Error: Go build failed."
    exit 1
fi
echo "Executable built: $EXECUTABLE_PATH"

# Generate .icns icon file
echo "Generating .icns icon from $ICON_PNG..."
ICONSET_DIR="$BUILD_DIR/$APP_NAME.iconset"
mkdir -p "$ICONSET_DIR"

# Check if icon.png exists
if [ ! -f "$ICON_PNG" ]; then
    echo "Error: Icon file not found at $ICON_PNG. Please ensure it exists."
    exit 1
fi

# Resize images and create .iconset
sips -z 16 16     "$ICON_PNG" --out "$ICONSET_DIR/icon_16x16.png"
sips -z 32 32     "$ICON_PNG" --out "$ICONSET_DIR/icon_16x16@2x.png"
sips -z 32 32     "$ICON_PNG" --out "$ICONSET_DIR/icon_32x32.png"
sips -z 64 64     "$ICON_PNG" --out "$ICONSET_DIR/icon_32x32@2x.png"
sips -z 128 128   "$ICON_PNG" --out "$ICONSET_DIR/icon_128x128.png"
sips -z 256 256   "$ICON_PNG" --out "$ICONSET_DIR/icon_128x128@2x.png"
sips -z 256 256   "$ICON_PNG" --out "$ICONSET_DIR/icon_256x256.png"
sips -z 512 512   "$ICON_PNG" --out "$ICONSET_DIR/icon_256x256@2x.png"
sips -z 512 512   "$ICON_PNG" --out "$ICONSET_DIR/icon_512x512.png"
sips -z 1024 1024 "$ICON_PNG" --out "$ICONSET_DIR/icon_512x512@2x.png"

# Convert .iconset to .icns
iconutil -c icns "$ICONSET_DIR" -o "$RESOURCES_DIR/AppIcon.icns"
if [ $? -ne 0 ]; then
    echo "Error: iconutil failed."
    exit 1
fi
echo "Icon generated: $RESOURCES_DIR/AppIcon.icns"

# Create Info.plist
echo "Creating Info.plist..."
APP_NAME_LOWER=$(echo "$APP_NAME" | tr '[:upper:]' '[:lower:]')
cat << EOF > "$APP_BUNDLE/Contents/Info.plist"
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
	<string>1.24.1</string>
	<key>LSUIElement</key>
	<true/>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>
EOF
if [ $? -ne 0 ]; then
    echo "Error: Failed to create Info.plist."
    exit 1
fi
echo "Info.plist created."

# Create PkgInfo
echo "Creating PkgInfo..."
echo "APPL????" > "$APP_BUNDLE/Contents/PkgInfo"
if [ $? -ne 0 ]; then
    echo "Error: Failed to create PkgInfo."
    exit 1
fi
echo "PkgInfo created."

# Clean up temporary iconset directory
echo "Cleaning up temporary iconset directory..."
rm -rf "$ICONSET_DIR"

# Touch the app bundle to refresh Finder metadata
echo "Touching app bundle to refresh Finder metadata..."
touch "$APP_BUNDLE"

echo "Build complete! Your application is located at $APP_BUNDLE"
echo "To run, navigate to the build directory and double-click CriptoMenu.app"
echo "If the icon does not appear, try running: killall Dock; killall Finder"