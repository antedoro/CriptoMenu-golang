# CriptoMenu

CriptoMenu is a simple macOS menubar application that allows you to monitor cryptocurrency quotes from Binance in real-time.

## Features

*   **Real-time Quotes:** Displays the price of a selected cryptocurrency pair directly in the menubar.
*   **Binance Support:** Connects to the Binance Spot API to fetch price data.
*   **Rounded Prices:** Prices are displayed rounded to two decimal places.
*   **Flexible Configuration:** Define the cryptocurrency pairs to monitor via a JSON configuration file.
*   **Interactive Menu:**
    *   **Monitored Pairs:** Select the pair to display on the fly from your configured list.
    *   **Edit Config:** Opens the `~/.criptomenu.json` configuration file in your default editor for easy modification.
    *   **Automatic Update:** The "Monitored Pairs" menu automatically updates when you save changes to the configuration file.
    *   **Quit:** Exits the application.
*   **Standalone Application:** Distributed as a native macOS `.app` application.

## Installation

1.  **Clone the repository (if available) or download the source files.**
2.  **Ensure you have Go installed.** You can download it from [go.dev](https://go.dev/dl/).
3.  **Build the application:**
    Open your terminal in the project's root directory and use the following commands:
    ```bash
    go build -o CriptoMenu.app/Contents/MacOS/CriptoMenu
    ```
4.  **Generate the application icon (.icns):**
    Make sure you have an `icon.png` file (ideally square, good quality) in your project directory. This script will create the `.icns` icon and place it in the app bundle:
    ```bash
    mkdir CriptoMenu.iconset
    sips -z 16 16     icon.png --out CriptoMenu.iconset/icon_16x16.png
    sips -z 32 32     icon.png --out CriptoMenu.iconset/icon_16x16@2x.png
    sips -z 32 32     icon.png --out CriptoMenu.iconset/icon_32x32.png
    sips -z 64 64     icon.png --out CriptoMenu.iconset/icon_32x32@2x.png
    sips -z 128 128   icon.png --out CriptoMenu.iconset/icon_128x128.png
    sips -z 256 256   icon.png --out CriptoMenu.iconset/icon_128x128@2x.png
    sips -z 256 256   icon.png --out CriptoMenu.iconset/icon_256x256.png
    sips -z 512 512   icon.png --out CriptoMenu.iconset/icon_256x256@2x.png
    sips -z 512 512   icon.png --out CriptoMenu.iconset/icon_512x512.png
    sips -z 1024 1024 icon.png --out CriptoMenu.iconset/icon_512x512@2x.png
    iconutil -c icns CriptoMenu.iconset
    mv CriptoMenu.icns CriptoMenu.app/Contents/Resources/AppIcon.icns
    rm -rf CriptoMenu.iconset
    ```
5.  **Create the `Info.plist` file:**
    This file defines the application's metadata. Create it as `CriptoMenu.app/Contents/Info.plist` with the following content:
    ```xml
    <?xml version="1.0" encoding="UTF-8"?>
    <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
    <plist version="1.0">
    <dict>
        <key>CFBundleExecutable</key>
        <string>CriptoMenu</string>
        <key>CFBundleIconFile</key>
        <string>AppIcon</string>
        <key>CFBundleIdentifier</key>
        <string>com.antedoro.criptomenu</string>
        <key>CFBundleName</key>
        <string>CriptoMenu</string>
        <key>CFBundlePackageType</key>
        <string>APPL</string>
        <key>CFBundleShortVersionString</key>
        <string>1.0.0</string>
        <key>LSUIElement</key>
        <true/>
        <key>NSHighResolutionCapable</key>
        <true/>
    </dict>
    </plist>
    ```
6.  **Create the `PkgInfo` file:**
    This is a small file for macOS. Create it as `CriptoMenu.app/Contents/PkgInfo` with the following content:
    ```
    APPL????
    ```
7.  **Move the application:**
    Move the `CriptoMenu.app` folder to your `/Applications` folder or any other desired location.

## Usage

1.  **Launch the application:** Double-click on `CriptoMenu.app`.
2.  **Configure monitored pairs:**
    *   Click on the application icon in the menubar.
    *   Select "Edit Config". This will open the `~/.criptomenu.json` file in your default editor.
    *   Modify the `"Pairs"` array with the cryptocurrency pairs you wish to monitor (e.g., `["BTCUSDC", "ETHUSDC", "BNBUSDT"]`).
    *   Save the file. The "Monitored Pairs" menu will update automatically.
3.  **Select the pair to display:**
    *   Click on the application icon in the menubar.
    *   Hover over "Monitored Pairs".
    *   Click on the pair you want to display in the menubar.

## Troubleshooting

*   **Icon not displayed correctly:** If the app icon doesn't appear or shows a generic icon, the system might have cached it. Try moving `CriptoMenu.app` to another folder and then back to its original location, or run the following command in the terminal:
    ```bash
    touch CriptoMenu.app; killall Dock; killall Finder
    ```
*   **Prices not updating / Errors:** Ensure you have an active internet connection. Verify that the pair symbols in `~/.criptomenu.json` are valid on Binance (e.g., `BTCUSDC`, not `BTC-USDC`).

## Technologies Used

*   **Go Lang:** The primary programming language.
*   **systray:** Library for managing the system tray icon and menu.
*   **Binance Connector Go:** Library for interacting with the Binance API.
*   **AppleScript (osascript):** Used to open the configuration file with the default editor.
