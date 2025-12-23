package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	binance_connector "github.com/binance/binance-connector-go"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
)

var (
	// Current monitored pair state
	currentPair      string
	currentPairMutex sync.RWMutex

	// Price Cache
	latestPrices      = make(map[string]float64)
	latestPricesMutex sync.RWMutex

	// Channel to trigger immediate price update
	updateChan = make(chan struct{}, 1)
)

// --- Core Logic ---

func rotatePairs() {
	log.Println("Pair rotation started.")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		configMutex.RLock()
		pairs := activeConfig.Pairs
	
pinned := activeConfig.PinnedPair
		configMutex.RUnlock()

		// Check if pinned
		if pinned != "" {
			setPair(pinned)
			
			// Update UI immediately from cache
			latestPricesMutex.RLock()
			price, ok := latestPrices[pinned]
			latestPricesMutex.RUnlock()

			if ok {
				roundedPrice := fmt.Sprintf("%.2f", price)
				systray.SetTitle(fmt.Sprintf("%s: %s", pinned, roundedPrice))
			} else {
				systray.SetTitle(fmt.Sprintf("%s: ...", pinned))
			}
			continue
		}

		if len(pairs) <= 1 {
			continue
		}

		current := getPair()
		nextIndex := 0
		found := false

		// Find current index
		for i, p := range pairs {
			if p == current {
				nextIndex = (i + 1) % len(pairs)
				found = true
				break
			}
		}
		
		// If current pair isn't in list (e.g. config changed), start at 0
		if !found {
			nextIndex = 0
		}

		nextPair := pairs[nextIndex]
		setPair(nextPair)

		// Update UI immediately from cache
		latestPricesMutex.RLock()
		price, ok := latestPrices[nextPair]
		latestPricesMutex.RUnlock()

		if ok {
			roundedPrice := fmt.Sprintf("%.2f", price)
			systray.SetTitle(fmt.Sprintf("%s: %s", nextPair, roundedPrice))
		} else {
			systray.SetTitle(fmt.Sprintf("%s: ...", nextPair))
		}
	}
}

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

	pairsToFetch := make(map[string]bool)

	configMutex.RLock()
	if activeConfig != nil {
		// Fetch ALL configured pairs (for rotation)
		for _, p := range activeConfig.Pairs {
			pairsToFetch[p] = true
		}
		// Fetch Alert pairs (for monitoring)
		for _, alert := range activeConfig.Alerts {
			if alert.Active {
				pairsToFetch[alert.Pair] = true
			}
		}
		// Fetch Pinned pair
		if activeConfig.PinnedPair != "" {
			pairsToFetch[activeConfig.PinnedPair] = true
		}
	}
	configMutex.RUnlock()

	if len(pairsToFetch) == 0 {
		return
	}

	// Fetch prices for all identified pairs
	for pair := range pairsToFetch {
		res, err := client.NewTickerPriceService().Symbol(pair).Do(context.Background())
		
		// Handle fetch error
		if err != nil {
			log.Printf("Error fetching %s price: %v", pair, err)
			continue
		}

		// Process response
		if len(res) > 0 {
			priceStr := res[0].Price
			priceFloat, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				log.Printf("Error parsing price string for %s: %v", pair, err)
				continue
			}

			// Update Cache
			latestPricesMutex.Lock()
			latestPrices[pair] = priceFloat
			latestPricesMutex.Unlock()

			// Update UI ONLY if this is the currently selected pair
			if pair == getPair() {
				roundedPrice := fmt.Sprintf("%.2f", priceFloat)
				systray.SetTitle(fmt.Sprintf("%s: %s", pair, roundedPrice))
				systray.SetTooltip(fmt.Sprintf("%s: %s", pair, roundedPrice))
			}

			// Check alerts for this pair (always, for background monitoring)
			checkAlerts(pair, priceFloat)
		}
	}
}

// checkAlerts iterates through configured alerts and triggers notifications if conditions are met.
func checkAlerts(pair string, price float64) {
	configMutex.Lock()
	defer configMutex.Unlock()

	cfg := activeConfig
	alertsChanged := false

	for _, alert := range cfg.Alerts {
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
					
					iconPath := "/Users/antedoro/Desktop/CriptoMenu-golang/icon.png"
					script := fmt.Sprintf(`
set iconPath to POSIX file "%s"
try
	display dialog "%s" with title "CriptoMenu Alert" buttons {"OK"} default button "OK" with icon iconPath
on error
	display alert "CriptoMenu Alert" message "%s"
end try`, iconPath, safeMsg, safeMsg)

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
			// cfg.Alerts[i].Active = false
			// alertsChanged = true
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

func setPair(pair string) {
	currentPairMutex.Lock()
	defer currentPairMutex.Unlock()
	currentPair = pair

	// Update Pin menu text
	if mPin != nil {
		configMutex.RLock()
		pinned := activeConfig.PinnedPair
		configMutex.RUnlock()

		if pinned == pair {
			mPin.SetTitle("Unpin " + pair)
		} else {
			mPin.SetTitle("Pin " + pair)
		}
	}
}

func getPair() string {
	currentPairMutex.RLock()
	defer currentPairMutex.RUnlock()
	return currentPair
}
