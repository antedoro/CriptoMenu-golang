package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	binance_connector "github.com/binance/binance-connector-go"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/pelletier/go-toml/v2"
)

// Alert struct defines a price alert condition
type Alert struct {
	ID        string  `toml:"id,omitempty"` // Optional identifier
	Pair      string  `toml:"pair"`
	Target    float64 `toml:"target"`
	Condition string  `toml:"condition"` // "above", "below"
	Active    bool    `toml:"active"`
}

// Config struct to hold application preferences
type Config struct {
	Pairs  []string `toml:"Pairs"`
	Alerts []Alert  `toml:"Alerts"`
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
	mEditConfig := systray.AddMenuItem("Edit Config", "Open config.toml")
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
	// Identify all unique pairs to fetch:
	// 1. The currently selected pair (for the menu bar display)
	// 2. Any pair that has an active alert (for background monitoring)
	pairsToFetch := make(map[string]bool)

	current := getPair()
	if current != "" {
		pairsToFetch[current] = true
	}

	configMutex.RLock()
	if activeConfig != nil {
		for _, alert := range activeConfig.Alerts {
			if alert.Active {
				pairsToFetch[alert.Pair] = true
			}
		}
	}
	configMutex.RUnlock()

	if len(pairsToFetch) == 0 {
		if current == "" {
			systray.SetTitle("No Pair")
			systray.SetTooltip("No Pair")
		}
		return
	}

	// Fetch prices for all identified pairs
	for pair := range pairsToFetch {
		res, err := client.NewTickerPriceService().Symbol(pair).Do(context.Background())
		
		// Handle fetch error
		if err != nil {
			log.Printf("Error fetching %s price: %v", pair, err)
			// Only update UI error if it's the currently displayed pair
			if pair == getPair() {
				systray.SetTitle("Error")
				systray.SetTooltip("Error")
			}
			continue
		}

		// Process response
		if len(res) > 0 {
			priceStr := res[0].Price
			priceFloat, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				log.Printf("Error parsing price string for %s: %v", pair, err)
				if pair == getPair() {
					systray.SetTitle(fmt.Sprintf("%s: Err", pair))
					systray.SetTooltip(fmt.Sprintf("%s: Err", pair))
				}
				continue
			}

			// Update UI ONLY if this is the currently selected pair
			if pair == getPair() {
				roundedPrice := fmt.Sprintf("%.2f", priceFloat)
				systray.SetTitle(fmt.Sprintf("%s: %s", pair, roundedPrice))
				systray.SetTooltip(fmt.Sprintf("%s: %s", pair, roundedPrice))
			}

			// Check alerts for this pair (always, for background monitoring)
			checkAlerts(pair, priceFloat)

		} else {
			// Empty response
			if pair == getPair() {
				systray.SetTitle("N/A")
				systray.SetTooltip("N/A")
			}
		}
	}
}

// checkAlerts iterates through configured alerts and triggers notifications if conditions are met.
func checkAlerts(pair string, price float64) {
	configMutex.Lock()
	defer configMutex.Unlock()

	cfg := activeConfig
	alertsChanged := false

	for i, alert := range cfg.Alerts {
		if !alert.Active {
			continue
		}
		if alert.Pair != pair {
			continue
		}

		triggered := false
		if alert.Condition == "above" && price >= alert.Target {
			triggered = true
		} else if alert.Condition == "below" && price <= alert.Target {
			triggered = true
		}

		if triggered {
			msg := fmt.Sprintf("%s ha raggiunto %.2f (Target: %.2f)", pair, price, alert.Target)
			log.Printf("ALERT TRIGGERED: %s", msg)
			
			// Fire notification
			// Icon path is empty to use default or system icon
			if runtime.GOOS == "darwin" {
				// Use osascript display alert (modal) for better visibility
				// Run in goroutine to not block price updates while waiting for user to dismiss
				go func(message string) {
					// Escape double quotes in the message to prevent script errors
					safeMsg := strings.ReplaceAll(message, "\"", "\\\"")
					script := fmt.Sprintf("display alert \"CriptoMenu Alert\" message \"%s\"", safeMsg)
					err := exec.Command("osascript", "-e", script).Run()
					if err != nil {
						log.Printf("Error sending macOS alert: %v", err)
					}
				}(msg)
			} else {
				err := beeep.Notify("CriptoMenu Alert", msg, "")
				if err != nil {
					log.Printf("Error sending notification: %v", err)
				}
			}

			// Deactivate alert
			cfg.Alerts[i].Active = false
			alertsChanged = true
		}
	}

	if alertsChanged {
		// Save config to persist the deactivated state
		err := saveConfigInternal(cfg)
		if err != nil {
			log.Printf("Error saving config after alert trigger: %v", err)
		}
	}
}

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
		// Check if file does not exist (or wrapped "no such file" error)
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			log.Println("Config file not found. Creating default with comments...")
			
			// Create the default file with comments
			if createErr := createDefaultConfig(); createErr != nil {
				log.Printf("Error creating default config: %v", createErr)
			}

			// Load it back
			cfg, err = loadConfig()
			if err != nil {
				log.Printf("Error loading newly created config: %v", err)
				cfg = &Config{Pairs: []string{"BTCUSDC", "ETHUSDC"}} // Fallback
			}
		} else if strings.Contains(err.Error(), "toml:") || strings.Contains(err.Error(), "decode") {
			// TOML syntax error
			log.Printf("Config file has invalid TOML, ignoring changes: %v", err)
			
			configMutex.RLock()
			hasConfig := activeConfig != nil
			configMutex.RUnlock()

			if hasConfig {
				return
			}
			// Startup with bad file -> Fallback
			cfg = &Config{Pairs: []string{"BTCUSDC", "ETHUSDC"}}
		} else {
			// Other errors
			log.Printf("Error loading config: %v", err)
			cfg = &Config{Pairs: []string{"BTCUSDC", "ETHUSDC"}}
		}
	}
	
	if cfg != nil {
		configMutex.Lock()
		activeConfig = cfg
		configMutex.Unlock()
	}
}

func createDefaultConfig() error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	defaultContent := `# Configuration for CriptoMenu
#
# Pairs: List of Binance trading pairs to display in the menu.
#        Example: ["BTCUSDC", "ETHUSDC"]
#
# Alerts: Define price alerts.
#   - pair: The trading pair to monitor.
#   - target: The price level to trigger the alert.
#   - condition: "above" (trigger when price goes above target) or "below" (trigger when price drops below target).
#   - active: Set to true to enable the alert. The app will set this to false after it triggers.

Pairs = [
    "BTCUSDC",
    "ETHUSDC",
    "ADAUSDC",
    "SOLUSDC",
    "LTCUSDC"
]

# Example Alert (Uncomment and modify to use)
# [[Alerts]]
#   pair = "BTCUSDC"
#   target = 100000.0
#   condition = "above" # "above" or "below"
#   active = true

# [[Alerts]]
#   pair = "ETHUSDC"
#   target = 10000.0
#   condition = "below" # "above" or "below"
#   active = true

# [[Alerts]]
#   pair = "LTCUSDC"
#   target = 50.0
#   condition = "below" # "above" or "below"
#   active = false
`
	return os.WriteFile(path, []byte(defaultContent), 0644)
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
		systray.SetTitle(fmt.Sprintf("%s: ...", selectedPair))
        
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
	// Check current working directory first (Development mode)
	localPath, _ := filepath.Abs(".criptomenu.toml")
	if _, err := os.Stat(localPath); err == nil {
		// Only log this once or it might be spammy, but for now it's helpful
		return localPath, nil
	}

	// Fallback to Home directory (Production/App mode)
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".criptomenu.toml"), nil
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
	err = toml.Unmarshal(file, &cfg)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config TOML: %w", err)
	}
	return &cfg, nil
}

func saveConfigInternal(cfg *Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("could not marshal config to TOML: %w", err)
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
