package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rsmacapinlac/earshot-tui/internal/platform"
)

type checkStatus int

const (
	checkPending checkStatus = iota
	checkRunning
	checkPassed
	checkFailed
)

type preflightState int

const (
	preflightStateRunning preflightState = iota
	preflightStateFailed
	preflightStateDone
)

type preflightCheck struct {
	label  string
	status checkStatus
	detail string
}

// preflightModel runs a sequential list of startup checks and either
// auto-advances to the import screen or surfaces a specific actionable error.
//
// Checks (in order):
//  0. ffmpeg present
//  1. Python 3.10–3.12 found
//  2. Recordings directory writable
//  3. Processor venv ready
type preflightModel struct {
	deps    Deps
	state   preflightState
	checks  []preflightCheck
	spinner spinner.Model
	width   int
	height  int
}

func newPreflightModel(deps Deps) preflightModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return preflightModel{
		deps: deps,
		checks: []preflightCheck{
			{label: "ffmpeg present", status: checkRunning},
			{label: "Python 3.10+ found", status: checkPending},
			{label: "Recordings directory writable", status: checkPending},
			{label: "Processor venv ready", status: checkPending},
		},
		spinner: sp,
	}
}

func (m preflightModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runCheck(0))
}

func (m preflightModel) runCheck(index int) tea.Cmd {
	deps := m.deps
	return func() tea.Msg {
		switch index {
		case 0: // ffmpeg
			if _, err := exec.LookPath("ffmpeg"); err != nil {
				return preflightCheckMsg{
					index: 0,
					err: fmt.Errorf(
						"ffmpeg not found on PATH\n" +
							"     Install it with your system package manager:\n" +
							"       Ubuntu/Debian:  sudo apt install ffmpeg\n" +
							"       macOS:          brew install ffmpeg"),
				}
			}
			return preflightCheckMsg{index: 0}

		case 1: // Python
			resolver := platform.NewPythonResolver()
			if _, _, err := resolver.Find(); err != nil {
				return preflightCheckMsg{
					index: 1,
					err: fmt.Errorf(
						"Python 3.10+ not found\n" +
							"     Install it with your system package manager:\n" +
							"       Ubuntu/Debian:  sudo apt install python3\n" +
							"       macOS:          brew install python@3"),
				}
			}
			return preflightCheckMsg{index: 1}

		case 2: // Recordings directory
			dir := deps.RecordingsDir
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return preflightCheckMsg{
					index: 2,
					err:   fmt.Errorf("cannot create recordings directory %s: %v", dir, err),
				}
			}
			// Verify we can write to it.
			tmp, err := os.CreateTemp(dir, ".write-check-*")
			if err != nil {
				return preflightCheckMsg{
					index: 2,
					err:   fmt.Errorf("recordings directory %s is not writable: %v", dir, err),
				}
			}
			tmp.Close()
			os.Remove(tmp.Name())
			return preflightCheckMsg{index: 2}

		case 3: // Processor venv
			if err := deps.ProcManager.Setup(func(_ string) {}); err != nil {
				return preflightCheckMsg{
					index: 3,
					err:   fmt.Errorf("venv setup failed: %w", err),
				}
			}
			return preflightCheckMsg{index: 3}
		}
		return preflightDoneMsg{}
	}
}

func (m preflightModel) update(msg tea.Msg) (preflightModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if msg.String() == "q" && m.state == preflightStateFailed {
			return m, tea.Quit
		}

	case preflightCheckMsg:
		if msg.err != nil {
			m.checks[msg.index].status = checkFailed
			m.checks[msg.index].detail = msg.err.Error()
			m.state = preflightStateFailed
			return m, nil
		}
		m.checks[msg.index].status = checkPassed
		next := msg.index + 1
		if next >= len(m.checks) {
			m.state = preflightStateDone
			return m, func() tea.Msg { return preflightDoneMsg{} }
		}
		m.checks[next].status = checkRunning
		return m, m.runCheck(next)
	}
	return m, nil
}

func (m preflightModel) view() string {
	var b strings.Builder
	b.WriteString("\n  Preflight checks\n\n")

	for _, c := range m.checks {
		var marker string
		switch c.status {
		case checkPending:
			marker = styleMuted.Render("○")
		case checkRunning:
			marker = m.spinner.View()
		case checkPassed:
			marker = styleCompleted.Render("✓")
		case checkFailed:
			marker = styleError.Render("✗")
		}
		b.WriteString(fmt.Sprintf("  %s  %s\n", marker, c.label))
		if c.detail != "" {
			b.WriteString(fmt.Sprintf("\n     %s\n\n", styleError.Render(c.detail)))
		}
	}

	if m.state == preflightStateFailed {
		b.WriteString("\n  " + footerDivider + "\n")
		b.WriteString("  " + formatFooter("q", "quit") + "\n")
	}

	return b.String()
}
