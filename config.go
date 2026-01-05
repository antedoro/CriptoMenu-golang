package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	Pairs      []string `toml:"Pairs"`
	Alerts     []Alert  `toml:"Alerts"`
	PinnedPair string   `toml:"pinned_pair,omitempty"`
}

var (
	// Configuration state
	activeConfig *Config
	configMutex  sync.RWMutex
)

// --- Config Helpers ---

func getConfigFilePath() (string, error) {
	// 1. Check CWD (works for 'go run' and terminal launch)
	localPath, _ := filepath.Abs(".criptomenu.toml")
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	// 2. Check relative to Executable (Traverse up)
	exePath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exePath)
		// Traverse up to 5 levels (covers Content/MacOS/Bundle/Build/ProjectRoot)
		for i := 0; i < 5; i++ {
			checkPath := filepath.Join(dir, ".criptomenu.toml")
			if _, err := os.Stat(checkPath); err == nil {
				return checkPath, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break // Hit root
			}
			dir = parent
		}
	}

	// 3. Fallback to Home directory (Production/App mode)
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
			errMsg := fmt.Sprintf("Config file has invalid TOML. Using default.\nError: %v", err)
			log.Printf(errMsg)
			showErrorAlert("Config Error", errMsg)
			
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
			errMsg := fmt.Sprintf("Error loading config. Using default.\nError: %v", err)
			log.Printf(errMsg)
			showErrorAlert("Config Error", errMsg)

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
