package platform

import (
	"os"
	"path/filepath"
)

const appName = "earshot-tui"

type appDirs struct{}

// NewAppDirs returns an AppDirs backed by platform-standard directories.
func NewAppDirs() AppDirs {
	return &appDirs{}
}

func (d *appDirs) Config() string {
	base, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, appName)
}

func (d *appDirs) Cache() string {
	base, err := os.UserCacheDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".cache")
	}
	return filepath.Join(base, appName)
}

func (d *appDirs) Data() string {
	return platformDataDir()
}
