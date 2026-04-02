package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rsmacapinlac/earshot-tui/internal/config"
)

type activeScreen int

const (
	screenSetup activeScreen = iota
	screenPreflight
	screenImport
	screenLibrary
)

// Root is the top-level Bubble Tea model. It owns all sub-models and routes
// messages to whichever screen is currently active.
type Root struct {
	deps      Deps
	active    activeScreen
	setup     setupModel
	preflight preflightModel
	imp       importModel
	library   libraryModel
	width     int
	height    int
}

// NewRoot creates the root model and decides the opening screen:
//   - setup wizard  → if config has no device sources
//   - preflight     → if config is valid and not in stub mode
//   - import screen → if config is valid and stub mode is active
func NewRoot(deps Deps) *Root {
	r := &Root{deps: deps}
	if config.NeedsSetup(deps.Config) {
		r.active = screenSetup
		r.setup = newSetupModel(deps)
	} else if os.Getenv("EARSHOT_PROCESSOR_STUB") == "1" {
		r.active = screenImport
		r.imp = newImportModel(deps)
	} else {
		r.active = screenPreflight
		r.preflight = newPreflightModel(deps)
	}
	r.library = newLibraryModel(deps)
	return r
}

// Init satisfies tea.Model.
func (r *Root) Init() tea.Cmd {
	switch r.active {
	case screenSetup:
		return r.setup.Init()
	case screenPreflight:
		return r.preflight.Init()
	case screenImport:
		return r.imp.Init()
	case screenLibrary:
		return r.library.Init()
	}
	return nil
}

// Update satisfies tea.Model.
func (r *Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return r, tea.Quit
		}

	// ---- screen transitions -------------------------------------------------

	case setupCompleteMsg:
		r.deps.Config = msg.cfg
		if os.Getenv("EARSHOT_PROCESSOR_STUB") == "1" {
			r.imp = newImportModel(r.deps)
			r.active = screenImport
			return r, r.imp.Init()
		}
		r.preflight = newPreflightModel(r.deps)
		r.active = screenPreflight
		return r, r.preflight.Init()

	case preflightDoneMsg:
		r.imp = newImportModel(r.deps)
		r.active = screenImport
		return r, r.imp.Init()

	case switchToImportMsg:
		r.imp = newImportModel(r.deps)
		r.active = screenImport
		return r, r.imp.Init()

	case switchToLibraryMsg:
		r.library = newLibraryModel(r.deps)
		r.active = screenLibrary
		return r, r.library.Init()
	}

	// ---- delegate to active sub-model --------------------------------------
	var cmd tea.Cmd
	switch r.active {
	case screenSetup:
		r.setup, cmd = r.setup.update(msg)
	case screenPreflight:
		r.preflight, cmd = r.preflight.update(msg)
	case screenImport:
		r.imp, cmd = r.imp.update(msg)
	case screenLibrary:
		r.library, cmd = r.library.update(msg)
	}
	return r, cmd
}

// View satisfies tea.Model.
func (r *Root) View() string {
	switch r.active {
	case screenSetup:
		return r.setup.view()
	case screenPreflight:
		return r.preflight.view()
	case screenImport:
		return r.imp.view()
	case screenLibrary:
		return r.library.view()
	}
	return ""
}

// Satisfy the compiler — spinner.TickMsg is handled by sub-models.
var _ spinner.TickMsg = spinner.TickMsg{}
