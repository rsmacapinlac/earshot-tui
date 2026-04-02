//go:build darwin

package platform

import (
	"os"
	"path/filepath"
)

func platformDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Application Support", appName)
}
