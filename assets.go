package main

import (
	"encoding/base64"
	"log"
)

// --- Assets ---

func getIcon() []byte {
	iconBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="
	decoded, err := base64.StdEncoding.DecodeString(iconBase64)
	if err != nil {
		log.Fatal("Error decoding base64 icon:", err)
	}
	return decoded
}
