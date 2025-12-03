# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2025-12-03

### Changed

- Enhanced monitoring: The application now fetches prices for all unique pairs specified in the active alerts list, in addition to the currently selected pair, enabling comprehensive background monitoring of alert conditions.
- Improved macOS alert notifications: Switched from subtle notification banners to a modal `display alert` dialog for critical price alerts on macOS, ensuring visibility even when notifications are otherwise suppressed.

### Fixed

- Resolved an issue where `~/.criptomenu.json` could be overwritten with default settings if a JSON parsing error occurred, preventing loss of user configurations.

## [1.1.0] - 2025-12-03

### Added

- Added "Market Chart" menu item to open Binance trade page for the currently selected pair.
- Added "About" menu item to open the GitHub project page.

### Changed

- Reordered menu items in the menubar.
- Implemented immediate price update when a new pair is selected.

### Fixed

- Corrected the icon path in the `build_macos.sh` script.

## [1.0.0] - 2025-12-02

### Added

-   Initial Go application for macOS menubar.
-   Connection to Binance API to fetch cryptocurrency prices.
-   Display of selected cryptocurrency quotes in the menubar.
-   "Monitored Pairs" submenu to dynamically select which pair to display.
-   "Edit Config" menu item to open `~/.criptomenu.json` for configuration.
-   Automatic update of "Monitored Pairs" menu when `~/.criptomenu.json` changes.
-   Rounding of displayed prices to two decimal places.
-   "Quit" menu item to exit the application gracefully.
-   Standalone macOS `.app` bundle (`CriptoMenu.app`) with custom icon.
-   Configuration file now saved in user's home directory (`~/.criptomenu.json`) for portability.
-   `README.md` file with detailed instructions.

### Changed

-   Application name changed from `BinanceQuotations` to `CriptoMenu`.
-   Menubar tooltip updated to "CriptoMenu".

### Fixed

-   Resolved immediate application exit on startup.
-   Resolved app icon not appearing due to Finder caching issues.