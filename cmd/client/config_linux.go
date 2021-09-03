// +build darwin linux

package main

import (
	"os"
	"path/filepath"
)

func getConfigDirectory() string {
	osConfigDir := os.Getenv("HOME")
	os.UserConfigDir()
	return filepath.Join(osConfigDir, "fileshare-client")
}
