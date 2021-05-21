// +build darwin linux

package main

func getConfigDirectory() string {
	osConfigDir := os.Getenv("HOME")
	return filepath.Join(osConfigDir, "fileshare-client")
}
