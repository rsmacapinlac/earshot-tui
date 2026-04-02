package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rsmacapinlac/earshot-tui/internal/config"
	"github.com/rsmacapinlac/earshot-tui/internal/platform"
)

// setupScanDoneMsg is the result of the background device scan.
type setupScanDoneMsg struct {
	found *platform.MountedDevice
	err   error
}

// focusedField tracks which input has focus on the single setup screen.
type focusedField int

const (
	fieldName focusedField = iota
	fieldPath // only shown when device not auto-detected
)

// setupModel is the Bubble Tea model for the first-time setup wizard.
// All fields are presented on a single screen.
type setupModel struct {
	deps      Deps
	found     *platform.MountedDevice // non-nil when auto-detected
	nameInput textinput.Model
	pathInput textinput.Model
	focus     focusedField
	errMsg    string
	width     int
}

func newSetupModel(deps Deps) setupModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Device name"
	nameInput.CharLimit = 64
	nameInput.Focus()

	pathInput := textinput.New()
	pathInput.Placeholder = "/path/to/mount"
	pathInput.CharLimit = 256

	return setupModel{
		deps:      deps,
		nameInput: nameInput,
		pathInput: pathInput,
		focus:     fieldName,
	}
}

func (m setupModel) Init() tea.Cmd {
	return tea.Batch(doSetupScan(m.deps.Scanner), textinput.Blink)
}

func doSetupScan(scanner platform.MountScanner) tea.Cmd {
	return func() tea.Msg {
		devices, err := scanner.Scan()
		if err != nil {
			return setupScanDoneMsg{err: err}
		}
		if len(devices) == 0 {
			return setupScanDoneMsg{}
		}
		d := devices[0]
		return setupScanDoneMsg{found: &d}
	}
}

func (m setupModel) fields() []focusedField {
	if m.found != nil {
		return []focusedField{fieldName}
	}
	return []focusedField{fieldName, fieldPath}
}

func (m setupModel) update(msg tea.Msg) (setupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case setupScanDoneMsg:
		if msg.found != nil {
			m.found = msg.found
			m.nameInput.SetValue(m.found.Name)
		}
		// Focus stays on fieldName regardless.

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			return m.advanceFocus(1)
		case "shift+tab", "up":
			return m.advanceFocus(-1)
		case "enter":
			fields := m.fields()
			if m.focus == fields[len(fields)-1] {
				return m.confirm()
			}
			return m.advanceFocus(1)
		}
	}

	// Forward to the focused input.
	var cmd tea.Cmd
	switch m.focus {
	case fieldName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case fieldPath:
		m.pathInput, cmd = m.pathInput.Update(msg)
	}
	return m, cmd
}

func (m setupModel) advanceFocus(dir int) (setupModel, tea.Cmd) {
	fields := m.fields()
	idx := 0
	for i, f := range fields {
		if f == m.focus {
			idx = i
			break
		}
	}
	next := (idx + dir + len(fields)) % len(fields)
	m.focus = fields[next]

	m.nameInput.Blur()
	m.pathInput.Blur()
	switch m.focus {
	case fieldName:
		m.nameInput.Focus()
	case fieldPath:
		m.pathInput.Focus()
	}
	return m, textinput.Blink
}

func (m setupModel) confirm() (setupModel, tea.Cmd) {
	name := strings.TrimSpace(m.nameInput.Value())

	mountPath := ""
	if m.found != nil {
		mountPath = m.found.MountPath
	} else {
		mountPath = strings.TrimSpace(m.pathInput.Value())
	}

	if name == "" || mountPath == "" {
		m.errMsg = "All fields are required."
		return m, nil
	}
	if !platform.IsEarshotDevice(mountPath) {
		m.errMsg = fmt.Sprintf(
			"%s does not look like an earshot device (no recording folders found).", mountPath)
		return m, nil
	}

	cfg := &config.Config{
		DeviceSources: map[string]string{name: mountPath},
		RecordingsDir: m.deps.RecordingsDir,
	}
	configDir := m.deps.AppDirs.Config()
	if err := config.Save(configDir, cfg); err != nil {
		m.errMsg = fmt.Sprintf("Could not write %s: %v",
			filepath.Join(configDir, "config.json"), err)
		return m, nil
	}
	return m, func() tea.Msg { return setupCompleteMsg{cfg: cfg} }
}

func (m setupModel) view() string {
	var b strings.Builder
	b.WriteString("\n  Welcome to earshot-tui.\n\n")

	if m.found != nil {
		b.WriteString("  Earshot device found.\n\n")
		b.WriteString("  Name:             " + m.nameInput.View() + "\n")
		b.WriteString("  Path:             " + m.found.MountPath + "\n")
	} else {
		b.WriteString("  No earshot device detected.\n\n")
		b.WriteString("  Name:             " + m.nameInput.View() + "\n")
		b.WriteString("  Path:             " + m.pathInput.View() + "\n")
	}

	if m.errMsg != "" {
		b.WriteString("\n  " + styleError.Render(m.errMsg) + "\n")
	}

	b.WriteString("\n  " + footerDivider + "\n")
	b.WriteString("  " + formatFooter("tab", "next field", "enter", "confirm") + "\n")

	return b.String()
}
