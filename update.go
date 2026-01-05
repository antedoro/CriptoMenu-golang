package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

const (
	CurrentVersion = "1.24.1"
)

// GitHubRelease struct to parse release info
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// --- Update Logic ---

func checkForUpdates() {
	log.Println("Checking for updates...")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/antedoro/CriptoMenu-golang/releases/latest")
	if err != nil {
		log.Printf("Error checking for updates: %v", err)
		showErrorAlert("Update Check Error", "Could not connect to GitHub.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("GitHub API returned status: %s", resp.Status)
		showErrorAlert("Update Check Error", "GitHub API returned error: "+resp.Status)
		return
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		log.Printf("Error parsing update response: %v", err)
		showErrorAlert("Update Check Error", "Could not parse update information.")
		return
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	log.Printf("Latest version: %s, Current version: %s", latestVersion, CurrentVersion)

	if latestVersion != CurrentVersion {
		msg := fmt.Sprintf("A new version (%s) is available!\nCurrent: %s", latestVersion, CurrentVersion)
		if runtime.GOOS == "darwin" {
			go func() {
				safeMsg := strings.ReplaceAll(msg, "\"", "\\\"")
				script := fmt.Sprintf(`display dialog "%s" with title "Update Available" buttons {"Download", "Cancel"} default button "Download"`, safeMsg)
				out, err := exec.Command("osascript", "-e", script).Output()
				if err == nil && strings.Contains(string(out), "button returned:Download") {
					_ = exec.Command("open", release.HTMLURL).Run()
				}
			}()
		} else {
			// Fallback for non-macOS (though this is mac-centric app)
			beeep.Notify("Update Available", msg, "")
			_ = exec.Command("open", release.HTMLURL).Run()
		}
	} else {
		msg := fmt.Sprintf("You are using the latest version (%s).", CurrentVersion)
		if runtime.GOOS == "darwin" {
			go func() {
				safeMsg := strings.ReplaceAll(msg, "\"", "\\\"")
				script := fmt.Sprintf(`display dialog "%s" with title "No Update Available" buttons {"OK"} default button "OK"`, safeMsg)
				_ = exec.Command("osascript", "-e", script).Run()
			}()
		}
	}
}
