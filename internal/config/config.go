// Package config manages the earshot-tui configuration file.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all user-editable settings.
type Config struct {
	DeviceSources map[string]string `json:"device_sources"`

	// RecordingsDir is the local directory where imported recordings are stored.
	// Defaults to {AppDirs.Data}/recordings/ when empty or absent.
	// Linux default:  ~/.local/share/earshot-tui/recordings/
	// macOS default:  ~/Library/Application Support/earshot-tui/recordings/
	RecordingsDir string `json:"recordings_dir,omitempty"`

}

// Load reads config.json from configDir. Returns an empty Config (not an error)
// when the file does not exist — the app should then start the setup wizard.
func Load(configDir string) (*Config, error) {
	path := filepath.Join(configDir, "config.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &cfg, nil
}

// Save writes cfg atomically (write-then-rename) to configDir/config.json.
func Save(configDir string, cfg *Config) error {
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create config dir %s: %w", configDir, err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	dest := filepath.Join(configDir, "config.json")
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename %s → %s: %w", tmp, dest, err)
	}
	return nil
}

// NeedsSetup reports whether cfg requires the setup wizard to run.
func NeedsSetup(cfg *Config) bool {
	return len(cfg.DeviceSources) == 0
}

// FirstDevice returns the name and path of the first configured device source,
// or empty strings if none are configured.
func FirstDevice(cfg *Config) (name, path string) {
	for n, p := range cfg.DeviceSources {
		return n, p
	}
	return "", ""
}

// EffectiveRecordingsDir returns the recordings directory to use.
// If cfg.RecordingsDir is set it is used directly; otherwise the default
// platform data directory is used.
func EffectiveRecordingsDir(cfg *Config, dataDir string) string {
	if cfg.RecordingsDir != "" {
		return cfg.RecordingsDir
	}
	return filepath.Join(dataDir, "recordings")
}
