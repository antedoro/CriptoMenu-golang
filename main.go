package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	binance_connector "github.com/binance/binance-connector-go"
	"github.com/getlantern/systray"
)

// ... (Rest of the structs and vars remain the same) ...

// Config struct to hold application preferences
type Config struct {
	Pairs []string `json:"Pairs"`
}

var (
	// Current monitored pair state
	currentPair      string
	currentPairMutex sync.RWMutex

	// Configuration state
	activeConfig *Config
	configMutex  sync.RWMutex

	// Menu items state
	mPairs        *systray.MenuItem
	pairMenuItems []*systray.MenuItem
    
    // Channel to trigger immediate price update
    updateChan = make(chan struct{}, 1)
)

func main() {
	systray.Run(onReady, onExit)
}

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
	mEditConfig := systray.AddMenuItem("Edit Config", "Open config.json")
	go func() {
		for range mEditConfig.ClickedCh {
			configPath, _ := getConfigFilePath()
			log.Printf("Opening config file at %s...", configPath)
			
			cmd := exec.Command("open", configPath)
			err := cmd.Run()
			if err != nil {
				log.Printf("Error opening config file: %v", err)
			}
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

	log.Println("onReady finished.")
}

func onExit() {
	log.Println("Application exiting.")
}

// --- Core Logic ---

func fetchPrices() {
	log.Println("Price fetching goroutine started.")
	client := binance_connector.NewClient("", "", "https://api.binance.com")
	
    // Initial update
    updatePrice(client)

    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

	for {
        select {
        case <-ticker.C:
            updatePrice(client)
        case <-updateChan:
            log.Println("Immediate price update triggered.")
            updatePrice(client)
            ticker.Reset(30 * time.Second) // Reset ticker to avoid double update
        }
	}
}

func updatePrice(client *binance_connector.Client) {
    pair := getPair()
    if pair == "" {
        systray.SetTitle("No Pair")
        systray.SetTooltip("No Pair")
        return
    }

    res, err := client.NewTickerPriceService().Symbol(pair).Do(context.Background())
    if err != nil {
        log.Printf("Error fetching %s price: %v", pair, err)
        if getPair() == pair {
            systray.SetTitle("Error")
            systray.SetTooltip("Error")
        }
    } else if len(res) > 0 {
        priceStr := res[0].Price
        priceFloat, err := strconv.ParseFloat(priceStr, 64)
        if err != nil {
            log.Printf("Error parsing price string to float: %v", err)
            if getPair() == pair {
                systray.SetTitle(fmt.Sprintf("%s: Err", pair))
                systray.SetTooltip(fmt.Sprintf("%s: Err", pair))
            }
        } else {
            roundedPrice := fmt.Sprintf("%.2f", priceFloat)
            if getPair() == pair {
                systray.SetTitle(fmt.Sprintf("%s: %s", pair, roundedPrice))
                systray.SetTooltip(fmt.Sprintf("%s: %s", pair, roundedPrice))
            }
        }
    } else {
        if getPair() == pair {
            systray.SetTitle("N/A")
            systray.SetTooltip("N/A")
        }
    }
}

// updateDisplay was removed.
// The systray.SetTooltip("CriptoMenu") in onReady is sufficient as a base.

func watchConfig() {
	ticker := time.NewTicker(2 * time.Second)
	var lastModTime time.Time
	configPath, err := getConfigFilePath()
	if err != nil {
		log.Printf("Error getting config path for watcher: %v", err)
		return
	}

	for range ticker.C {
		info, err := os.Stat(configPath)
		if err != nil {
			continue
		}
		if !lastModTime.IsZero() && info.ModTime() != lastModTime {
			log.Println("Config file changed. Reloading...")
			loadAndSetConfig()
			updatePairsMenu()
		}
		lastModTime = info.ModTime()
	}
}

func loadAndSetConfig() {
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("Error loading config, using default: %v", err)
		cfg = &Config{Pairs: []string{"BTCUSDC", "ETHUSDC"}}
		_ = saveConfig(cfg)
	}
	configMutex.Lock()
	activeConfig = cfg
	configMutex.Unlock()
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

	// Update existing items
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
		systray.SetTitle(fmt.Sprintf("%s: ...", selectedPair)) // Reverted to original
        
        // Trigger immediate update
        select {
        case updateChan <- struct{}{}:
        default:
            // Channel full, update already pending
        }
	}
}

func setPair(pair string) {
	currentPairMutex.Lock()
	defer currentPairMutex.Unlock()
	currentPair = pair
}

func getPair() string {
	currentPairMutex.RLock()
	defer currentPairMutex.RUnlock()
	return currentPair
}

// --- Config Helpers ---

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".criptomenu.json"), nil
}

func loadConfig() (*Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config JSON: %w", err)
	}
	return &cfg, nil
}

func saveConfig(cfg *Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal config to JSON: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}
	return nil
}

// --- Assets ---

func getIcon() []byte {
	iconBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="
	decoded, err := base64.StdEncoding.DecodeString(iconBase64)
	if err != nil {
		log.Fatal("Error decoding base64 icon:", err)
	}
	return decoded
}

