package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rsmacapinlac/earshot-tui/internal/config"
	"github.com/rsmacapinlac/earshot-tui/internal/platform"
	"github.com/rsmacapinlac/earshot-tui/internal/processor"
	"github.com/rsmacapinlac/earshot-tui/internal/recording"
	"github.com/rsmacapinlac/earshot-tui/internal/tui"
)

func main() {
	dirs := platform.NewAppDirs()

	// Load or create empty config (missing file is not an error).
	cfg, err := config.Load(dirs.Config())
	if err != nil {
		fmt.Fprintf(os.Stderr, "earshot-tui: cannot read config: %v\n", err)
		os.Exit(1)
	}

	// Resolve the recordings directory once: config override or platform default.
	recordingsDir := config.EffectiveRecordingsDir(cfg, dirs.Data())

	// Interrupt recovery: reset any stuck "processing" recordings to "interrupted".
	if err := recording.RecoverInterrupted(recordingsDir); err != nil {
		log.Printf("interrupt recovery: %v", err)
	}

	// Locate Python (skip in stub mode — Python not required for TUI testing).
	var pythonPath string
	if os.Getenv("EARSHOT_PROCESSOR_STUB") != "1" && !config.NeedsSetup(cfg) {
		resolver := platform.NewPythonResolver()
		var pyErr error
		pythonPath, _, pyErr = resolver.Find()
		if pyErr != nil {
			fmt.Fprintf(os.Stderr, "earshot-tui: %v\n", pyErr)
			os.Exit(1)
		}
	}

	procManager := processor.NewManager(pythonPath, dirs.Config(), dirs.Cache())

	deps := tui.Deps{
		AppDirs:       dirs,
		Config:        cfg,
		RecordingsDir: recordingsDir,
		Scanner:       platform.NewMountScanner(),
		Player:        platform.NewAudioPlayer(),
		ProcManager:   procManager,
	}

	root := tui.NewRoot(deps)
	p := tea.NewProgram(root, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "earshot-tui: %v\n", err)
		os.Exit(1)
	}
}
