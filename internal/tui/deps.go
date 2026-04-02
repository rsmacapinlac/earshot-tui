package tui

import (
	"github.com/rsmacapinlac/earshot-tui/internal/config"
	"github.com/rsmacapinlac/earshot-tui/internal/platform"
	"github.com/rsmacapinlac/earshot-tui/internal/processor"
)

// Deps bundles all external dependencies the TUI needs.
// Constructed in main and passed to NewRoot.
type Deps struct {
	AppDirs       platform.AppDirs
	Config        *config.Config
	RecordingsDir string // effective recordings directory (config override or platform default)
	Scanner       platform.MountScanner
	Player        platform.AudioPlayer
	ProcManager   *processor.Manager
}
