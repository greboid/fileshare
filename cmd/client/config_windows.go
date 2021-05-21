// +build windows

package main

import (
	"os"
	"path/filepath"
)

func getConfigDirectory() string {
	osConfigDir := os.Getenv("APPDATA")
	return filepath.Join(osConfigDir, "fileshare-client")
}
