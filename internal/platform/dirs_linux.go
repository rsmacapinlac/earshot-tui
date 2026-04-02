//go:build linux

package platform

import (
	"os"
	"path/filepath"
)

func platformDataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, appName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", appName)
}
