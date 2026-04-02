package config_test

import (
	"path/filepath"
	"testing"

	"github.com/rsmacapinlac/earshot-tui/internal/config"
)

func TestLoadMissing(t *testing.T) {
	dir := t.TempDir()
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !config.NeedsSetup(cfg) {
		t.Error("NeedsSetup should be true when config file is absent")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		DeviceSources: map[string]string{"Pi4-Earshot": "/run/media/ritchie/EARSHOT"},
	}
	if err := config.Save(dir, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if config.NeedsSetup(loaded) {
		t.Error("NeedsSetup should be false after saving a complete config")
	}
	if got := loaded.DeviceSources["Pi4-Earshot"]; got != "/run/media/ritchie/EARSHOT" {
		t.Errorf("DeviceSources[Pi4-Earshot] = %q, want %q", got, "/run/media/ritchie/EARSHOT")
	}
}

func TestNeedsSetupMissingDevice(t *testing.T) {
	cfg := &config.Config{}
	if !config.NeedsSetup(cfg) {
		t.Error("NeedsSetup should be true when DeviceSources is empty")
	}
}

func TestSaveAtomicOnBadDir(t *testing.T) {
	err := config.Save("/dev/null/notadir", &config.Config{})
	if err == nil {
		t.Error("expected an error saving into a non-existent nested path")
	}
}

func TestEffectiveRecordingsDir(t *testing.T) {
	dataDir := "/data"

	// No override — use platform default.
	cfg := &config.Config{}
	got := config.EffectiveRecordingsDir(cfg, dataDir)
	want := filepath.Join(dataDir, "recordings")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// With override.
	cfg.RecordingsDir = "/custom/recordings"
	got = config.EffectiveRecordingsDir(cfg, dataDir)
	if got != "/custom/recordings" {
		t.Errorf("got %q, want %q", got, "/custom/recordings")
	}
}
