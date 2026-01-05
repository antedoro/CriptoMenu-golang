package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/getlantern/systray"
)

var (
	// Menu items state
	mPairs        *systray.MenuItem
	mPin          *systray.MenuItem // New Pinned Item
	pairMenuItems []*systray.MenuItem
)

func onReady() {
	log.Println("onReady started.")
	systray.SetIcon(getIcon())
	systray.SetTitle("Loading...")
	systray.SetTooltip("CriptoMenu")

	// Initialize config
	loadAndSetConfig()

	// Set initial monitored pair
	configMutex.RLock()
	if len(activeConfig.Pairs) > 0 {
		setPair(activeConfig.Pairs[0])
	} else {
		setPair("BTCUSDC") // Fallback
	}
	configMutex.RUnlock()

	// "Monitored Pairs" Parent Menu
	mPairs = systray.AddMenuItem("Monitored Pairs", "Select a pair to display")

	// Initialize the submenus based on current config
	updatePairsMenu()

	// "Pin/Unpin" menu item
	mPin = systray.AddMenuItem("Pin Current Pair", "Fix the current pair to the menu bar")
	go func() {
		for range mPin.ClickedCh {
			configMutex.Lock()
			current := getPair()
			
			if activeConfig.PinnedPair == current {
				// Unpin
				activeConfig.PinnedPair = ""
				mPin.SetTitle("Pin " + current)
			} else {
				// Pin
				activeConfig.PinnedPair = current
				mPin.SetTitle("Unpin " + current)
			}
			
			// Save config
			err := saveConfigInternal(activeConfig)
			configMutex.Unlock()
			
			if err != nil {
				log.Printf("Error saving config after pin/unpin: %v", err)
			} else {
				// Force UI update
				updatePairsMenu() // Refresh checkmarks if implemented, or just state
			}
		}
	}()

	// "Market Chart" menu item
	mMarketChart := systray.AddMenuItem("Market Chart", "Open Binance chart for current pair")
	go func() {
		for range mMarketChart.ClickedCh {
			pair := getPair()
			if pair == "" {
				continue
			}
			// Simple heuristic to split pair for URL: BTCUSDC -> BTC_USDC
			// Common quote currencies
			quotes := []string{"USDT", "USDC", "BUSD", "EUR", "BTC", "ETH", "BNB"}
			formattedPair := pair
			for _, q := range quotes {
				if strings.HasSuffix(pair, q) && len(pair) > len(q) {
					formattedPair = pair[:len(pair)-len(q)] + "_" + q
					break
				}
			}

			url := fmt.Sprintf("https://www.binance.com/it/trade/%s?type=spot", formattedPair)
			log.Printf("Opening chart: %s", url)
			_ = exec.Command("open", url).Run()
		}
	}()

	// "Edit Config" menu item
	mEditConfig := systray.AddMenuItem("Edit Config", "Open config.toml")
	
	// Update tooltip with actual path
	if path, err := getConfigFilePath(); err == nil {
		mEditConfig.SetTooltip(fmt.Sprintf("Editing: %s", path))
	}

	go func() {
		for range mEditConfig.ClickedCh {
			configPath, _ := getConfigFilePath()
			log.Printf("Opening config file at %s...", configPath)

			cmd := exec.Command("open", "-t", configPath)
			err := cmd.Run()
			if err != nil {
				log.Printf("Error opening config file: %v", err)
			}
		}
	}()

	// "Check for Update..." menu item
	mCheckUpdate := systray.AddMenuItem("Check for Update...", "Check for new releases on GitHub")
	go func() {
		for range mCheckUpdate.ClickedCh {
			checkForUpdates()
		}
	}()

	// "About" menu item
	mAbout := systray.AddMenuItem("About", "Open GitHub project page")
	go func() {
		for range mAbout.ClickedCh {
			_ = exec.Command("open", "https://github.com/antedoro/CriptoMenu-golang").Run()
		}
	}()

	systray.AddSeparator()

	// "Quit" menu item
	mQuit := systray.AddMenuItem("Quit", "Quit the app")
	go func() {
		<-mQuit.ClickedCh
		log.Println("Quit menu item clicked. Quitting systray.")
		systray.Quit()
	}()

	// Start file watcher for config changes
	go watchConfig()

	// Start price fetching
	go fetchPrices()

	// Start Pair Rotation (Carousel)
	go rotatePairs()

	log.Println("onReady finished.")
}

func onExit() {
	log.Println("Application exiting.")
}

func updatePairsMenu() {
	configMutex.RLock()
	pairs := activeConfig.Pairs
	configMutex.RUnlock()

	// Ensure we have enough menu items
	for i := len(pairMenuItems); i < len(pairs); i++ {
		// Create new item
		item := mPairs.AddSubMenuItem("", "")
		pairMenuItems = append(pairMenuItems, item)
		
		// Start listener for this specific index
		go func(index int, it *systray.MenuItem) {
			for range it.ClickedCh {
				handlePairClick(index)
			}
		}(i, item)
	}

	// Update existing items and hide excess ones
	for i, item := range pairMenuItems {
		if i < len(pairs) {
			item.SetTitle(pairs[i])
			item.SetTooltip("Display " + pairs[i])
			item.Show()
		} else {
			item.Hide()
		}
	}
}

func handlePairClick(index int) {
	configMutex.RLock()
	defer configMutex.RUnlock()
	
	if index >= 0 && index < len(activeConfig.Pairs) {
		selectedPair := activeConfig.Pairs[index]
		log.Printf("Selected pair: %s", selectedPair)
		setPair(selectedPair)
		systray.SetTitle(fmt.Sprintf("%s: ...", selectedPair))
        
        // Trigger immediate update
        select {
        case updateChan <- struct{}{}:
        default:
            // Channel full, update already pending
        }
	}
}

// Helper to show error alerts
func showErrorAlert(title, message string) {
	if runtime.GOOS == "darwin" {
		go func() {
			safeMsg := strings.ReplaceAll(message, "\"", "\\\"")
			script := fmt.Sprintf("display alert \"%s\" message \"%s\" as critical", title, safeMsg)
			_ = exec.Command("osascript", "-e", script).Run()
		}()
	}
}
