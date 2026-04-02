package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rsmacapinlac/earshot-tui/internal/processor"
	"github.com/rsmacapinlac/earshot-tui/internal/recording"
	"github.com/rsmacapinlac/earshot-tui/internal/transcript"
)

type libraryScreen int

const (
	libraryScreenLoading libraryScreen = iota
	libraryScreenIdle
	libraryScreenProcessing
)

type libraryEntry struct {
	rec      recording.Recording
	selected bool
	queued   bool
}

type libraryModel struct {
	deps         Deps
	screen       libraryScreen
	entries      []libraryEntry
	cursor       int
	activeDir    string
	processingCh <-chan processingProgressMsg // live channel; nil when idle
	progressBar  progress.Model
	spinner      spinner.Model
	errMsg       string
	width        int
	height       int
}

func newLibraryModel(deps Deps) libraryModel {
	bar := progress.New(progress.WithDefaultGradient())
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return libraryModel{
		deps:        deps,
		screen:      libraryScreenLoading,
		progressBar: bar,
		spinner:     sp,
	}
}

func (m libraryModel) Init() tea.Cmd {
	return tea.Batch(m.loadRecordings(), m.spinner.Tick)
}

func (m libraryModel) loadRecordings() tea.Cmd {
	recordingsDir := m.deps.RecordingsDir
	return func() tea.Msg {
		recs, err := recording.List(recordingsDir)
		return libraryLoadedMsg{recordings: recs, err: err}
	}
}

func (m libraryModel) update(msg tea.Msg) (libraryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progressBar.Width = m.width - 64

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		pm, cmd := m.progressBar.Update(msg)
		m.progressBar = pm.(progress.Model)
		return m, cmd

	case libraryLoadedMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.screen = libraryScreenIdle
			return m, nil
		}
		m.screen = libraryScreenIdle
		m.entries = make([]libraryEntry, len(msg.recordings))
		for i, r := range msg.recordings {
			m.entries[i] = libraryEntry{rec: r}
		}

	case processingProgressMsg:
		// Re-queue the drain and animate the bar.
		var cmds []tea.Cmd
		if msg.progress.Fraction >= 0 {
			cmds = append(cmds, m.progressBar.SetPercent(msg.progress.Fraction))
		}
		cmds = append(cmds, drainProcessing(m.processingCh))
		return m, tea.Batch(cmds...)

	case processingDoneMsg:
		m.processingCh = nil
		now := time.Now().UTC()
		for i := range m.entries {
			if m.entries[i].rec.Dir != msg.dir {
				continue
			}
			st := &m.entries[i].rec.Status
			if msg.err != nil {
				st.Status = recording.StateFailed
				st.Error = msg.err.Error()
			} else {
				st.Status = recording.StateCompleted
				st.TranscribedAt = &now
			}
			_ = recording.SaveStatus(msg.dir, st)
			m.entries[i].queued = false
			break
		}
		m.activeDir = ""
		return m.processNext()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m libraryModel) handleKey(msg tea.KeyMsg) (libraryModel, tea.Cmd) {
	switch m.screen {
	case libraryScreenIdle:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case " ":
			if len(m.entries) > 0 {
				e := &m.entries[m.cursor]
				if recording.IsRetryable(e.rec.Status.Status) {
					e.selected = !e.selected
				}
			}
		case "t":
			if m.hasSelected() {
				m.screen = libraryScreenProcessing
				for i := range m.entries {
					if m.entries[i].selected {
						m.entries[i].queued = true
						m.entries[i].selected = false
					}
				}
				return m.processNext()
			}
		case "enter":
			if len(m.entries) > 0 && m.entries[m.cursor].rec.Status.Status == recording.StateCompleted {
				return m, m.openTranscript(m.entries[m.cursor].rec.Dir)
			}
		case "b":
			return m, func() tea.Msg { return switchToImportMsg{} }
		case "q":
			return m, tea.Quit
		}

	case libraryScreenProcessing:
		if msg.String() == "c" {
			for i := range m.entries {
				if m.entries[i].rec.Dir == m.activeDir {
					m.entries[i].rec.Status.Status = recording.StateInterrupted
					_ = recording.SaveStatus(m.entries[i].rec.Dir, &m.entries[i].rec.Status)
				}
				if m.entries[i].queued {
					m.entries[i].queued = false
				}
			}
			m.screen = libraryScreenIdle
			m.activeDir = ""
			m.processingCh = nil
		}
	}
	return m, nil
}

func (m libraryModel) hasSelected() bool {
	for _, e := range m.entries {
		if e.selected {
			return true
		}
	}
	return false
}

// processNext starts transcribing the next queued recording using a channel-tap
// for live progress. Returns the updated model and initial Cmd.
func (m libraryModel) processNext() (libraryModel, tea.Cmd) {
	for i := range m.entries {
		if !m.entries[i].queued {
			continue
		}
		dir := m.entries[i].rec.Dir
		m.activeDir = dir
		m.entries[i].queued = false
		m.entries[i].rec.Status.Status = recording.StateProcessing
		_ = recording.SaveStatus(dir, &m.entries[i].rec.Status)

		mgr := m.deps.ProcManager
		audioPath := recording.AudioPath(dir)
		status := m.entries[i].rec.Status

		ch := make(chan processingProgressMsg, 16)
		m.processingCh = ch

		workCmd := func() tea.Msg {
			result, err := mgr.Run(audioPath, func(p processor.Progress) {
				ch <- processingProgressMsg{dir: dir, progress: p}
			})
			close(ch)
			if err != nil {
				return processingDoneMsg{dir: dir, err: err}
			}
			if werr := transcript.Write(dir, &status, result); werr != nil {
				return processingDoneMsg{dir: dir, err: werr}
			}
			return processingDoneMsg{dir: dir}
		}

		return m, tea.Batch(workCmd, drainProcessing(ch))
	}

	m.screen = libraryScreenIdle
	return m, nil
}

func (m libraryModel) openTranscript(dir string) tea.Cmd {
	tPath := recording.TranscriptPath(dir)
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, tPath)
	return tea.ExecProcess(cmd, func(_ error) tea.Msg {
		return m.loadRecordings()()
	})
}

func (m libraryModel) view() string {
	var b strings.Builder

	b.WriteString("\n  " + styleBold.Render("Library") + "\n\n")

	if m.screen == libraryScreenLoading {
		b.WriteString(fmt.Sprintf("\n  %s Loading library…\n", m.spinner.View()))
		return b.String()
	}

	if len(m.entries) == 0 {
		b.WriteString("\n  No recordings yet. Press [b] to go back and import from a device.\n")
		b.WriteString("\n  " + footerDivider + "\n")
		b.WriteString("  " + formatFooter("b", "back", "q", "quit") + "\n")
		return b.String()
	}

	if m.errMsg != "" {
		b.WriteString("  " + styleError.Render(m.errMsg) + "\n\n")
	}

	for i, e := range m.entries {
		cur := "  "
		if i == m.cursor {
			cur = "> "
		}
		marker, colorFn := statusMarker(e)
		label := formatRecordingLabel(e)
		if e.queued || e.rec.Status.Status == recording.StateProcessing {
			label = stylePending.Render(label)
		}
		if m.screen == libraryScreenProcessing && e.rec.Dir == m.activeDir {
			b.WriteString(fmt.Sprintf("%s%s %s  %s\n", cur, colorFn(marker), label, m.progressBar.View()))
		} else {
			b.WriteString(fmt.Sprintf("%s%s %s\n", cur, colorFn(marker), label))
		}
	}

	b.WriteString("\n  " + footerDivider + "\n")
	if m.screen == libraryScreenProcessing {
		b.WriteString("  " + formatFooter("c", "cancel") + "\n")
	} else {
		b.WriteString("  " + formatFooter("space", "select", "t", "transcribe") + "\n")
		b.WriteString("  " + footerDivider + "\n")
		b.WriteString("  " + formatFooter("b", "back", "q", "quit") + "\n")
	}

	return b.String()
}

func statusMarker(e libraryEntry) (string, func(...string) string) {
	identity := func(strs ...string) string {
		if len(strs) == 0 {
			return ""
		}
		return strs[0]
	}
	if e.selected {
		return "[✓]", identity
	}
	if e.queued {
		return "[✓]", stylePending.Render
	}
	switch e.rec.Status.Status {
	case recording.StateCompleted:
		return "[✓]", styleCompleted.Render
	case recording.StateFailed:
		return "[✗]", styleError.Render
	case recording.StateInterrupted:
		return "[!]", styleError.Render
	case recording.StateProcessing:
		return "[✓]", stylePending.Render
	default:
		return "[ ]", identity
	}
}

func formatRecordingLabel(e libraryEntry) string {
	s := e.rec.Status
	ts := s.RecordedAt.Format("2006-01-02 15:04")
	raw := string(s.Status)
	stateLabel := strings.ToUpper(raw[:1]) + raw[1:]
	switch {
	case e.queued:
		stateLabel = "Waiting…"
	case s.Status == recording.StateProcessing:
		stateLabel = "Processing…"
	}
	dur := ""
	if s.Duration > 0 {
		dur = "  " + transcript.FormatDuration(s.Duration)
	}
	return fmt.Sprintf("%s  %-13s%s", ts, stateLabel, dur)
}
