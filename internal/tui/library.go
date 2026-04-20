package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	btable "github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rsmacapinlac/earshot-tui/internal/calendar"
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
	rec       recording.Recording
	queued    bool
	enriching bool
}

type libraryModel struct {
	deps           Deps
	screen         libraryScreen
	entries        []libraryEntry
	table          btable.Model
	activeDir      string
	processingCh   <-chan processingProgressMsg // live channel; nil when idle
	activeFraction float64                      // 0.0–1.0 progress of active transcription
	activePhase    string                       // current phase label from processor
	spinner        spinner.Model
	errMsg         string
	width          int
	height         int
}

// libraryTableStyles returns styles with no cell padding and a simple cursor indicator.
func libraryTableStyles() btable.Styles {
	return btable.Styles{
		Header:   lipgloss.NewStyle().Bold(true),
		Cell:     lipgloss.NewStyle(),
		Selected: lipgloss.NewStyle(), // cursor shown via ">" prefix in Date column
	}
}

// libraryColumns returns the column definitions given the current terminal width.
// The Title column absorbs all remaining space.
//
// Column widths (no cell padding — styles have no Padding set):
//
//	cursor:    2   ">" or " "
//	date:     12   "2006-01-02  " (10 + 2 trailing spaces for separation)
//	time:      8   "15:04   "
//	dur:      12   "1h 23m 45s  "
//	status:   14   "Transcribing  " / "Transcribed  " / etc.
//	progress: 14   "[████████░░]  " during transcription, empty otherwise
//	title:   dynamic
func libraryColumns(termWidth int) []btable.Column {
	const curW, dateW, timeW, durW, statusW, progressW = 2, 12, 8, 12, 14, 14
	fixed := curW + dateW + timeW + durW + statusW + progressW
	titleW := termWidth - 2 - fixed // subtract 2 for the "  " view margin
	if titleW < 10 {
		titleW = 10
	}
	return []btable.Column{
		{Title: "", Width: curW},
		{Title: "Date", Width: dateW},
		{Title: "Time", Width: timeW},
		{Title: "Duration", Width: durW},
		{Title: "Status", Width: statusW},
		{Title: "Progress", Width: progressW},
		{Title: "Title", Width: titleW},
	}
}

// progressBar renders a plain-Unicode bar using █/░ block characters.
// width is the visible character width including brackets, e.g. "[████████░░]" is 12.
func progressBar(fraction float64, width int) string {
	barW := width - 2 // subtract brackets
	if barW <= 0 {
		return ""
	}
	filled := int(fraction*float64(barW) + 0.5)
	if filled > barW {
		filled = barW
	}
	if filled < 0 {
		filled = 0
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", barW-filled) + "]"
}

// libraryTableHeight returns the viewport height to pass to table.SetHeight.
// Chrome: 3 lines (title block) + 4 lines (footer block) = 7. SetHeight(h)
// sets viewport.Height = h - 1 (header row), so table.View() prints h lines.
func libraryTableHeight(termHeight int) int {
	h := termHeight - 7
	if h < 5 {
		h = 5
	}
	return h
}

func newLibraryModel(deps Deps, width, height int) libraryModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	// Strip conflicting default keybindings so our own key handlers can fire.
	km := btable.DefaultKeyMap()
	km.PageUp = key.NewBinding()
	km.PageDown = key.NewBinding()
	km.HalfPageUp = key.NewBinding()
	km.HalfPageDown = key.NewBinding()
	km.GotoTop = key.NewBinding()
	km.GotoBottom = key.NewBinding()

	t := btable.New(
		btable.WithColumns(libraryColumns(width)),
		btable.WithFocused(true),
		btable.WithHeight(libraryTableHeight(height)),
		btable.WithKeyMap(km),
		btable.WithStyles(libraryTableStyles()),
	)

	return libraryModel{
		deps:    deps,
		screen:  libraryScreenLoading,
		spinner: sp,
		table:   t,
		width:   width,
		height:  height,
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

// buildRows converts m.entries into table rows.  The first (cursor) column
// holds ">" for the selected row and "  " for all others.
func (m libraryModel) buildRows() []btable.Row {
	cursor := m.table.Cursor()
	rows := make([]btable.Row, len(m.entries))
	for i, e := range m.entries {
		s := e.rec.Status

		cur := " "
		if i == cursor {
			cur = ">"
		}

		date := s.RecordedAt.Format("2006-01-02")
		startTime := s.RecordedAt.Format("15:04")

		dur := "—"
		if s.Duration > 0 {
			dur = transcript.FormatDuration(s.Duration)
		}

		title := ""
		if e.enriching {
			title = "enriching…"
		} else if s.Title != "" {
			title = s.Title
		}

		isActive := m.screen == libraryScreenProcessing && e.rec.Dir == m.activeDir
		isQueued := e.queued || s.Status == recording.StateProcessing

		var statusStr, progressStr string
		if isActive {
			switch m.activePhase {
			case "loading":
				statusStr = "Loading…"
			case "transcribing":
				statusStr = "Transcribing"
				progressStr = progressBar(m.activeFraction, 12)
			case "complete":
				statusStr = "Completing…"
			default:
				statusStr = "Processing…"
			}
		} else {
			statusStr = formatStatusLabel(s.Status, isQueued)
		}

		rows[i] = btable.Row{cur, date, startTime, dur, statusStr, progressStr, title}
	}
	return rows
}

func (m libraryModel) update(msg tea.Msg) (libraryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetColumns(libraryColumns(m.width))
		m.table.SetHeight(libraryTableHeight(m.height))
		m.table.SetRows(m.buildRows())

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
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
		m.table.SetRows(m.buildRows())

	case processingProgressMsg:
		if msg.progress.Phase != "" {
			m.activePhase = msg.progress.Phase
		}
		if msg.progress.Fraction >= 0 {
			m.activeFraction = msg.progress.Fraction
		}
		m.table.SetRows(m.buildRows())
		var cmds []tea.Cmd
		cmds = append(cmds, drainProcessing(m.processingCh))
		return m, tea.Batch(cmds...)

	case processingDoneMsg:
		m.processingCh = nil
		now := recording.FlexTime{Time: time.Now().UTC()}
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

	case enrichDoneMsg:
		for i := range m.entries {
			if m.entries[i].rec.Dir != msg.dir {
				continue
			}
			m.entries[i].enriching = false
			if msg.err != nil {
				m.errMsg = msg.err.Error()
				break
			}
			now := recording.FlexTime{Time: time.Now().UTC()}
			st := &m.entries[i].rec.Status
			st.Title = msg.title
			st.Attendees = msg.attendees
			st.Description = msg.description
			st.EnrichedAt = &now
			_ = recording.SaveStatus(msg.dir, st)
			break
		}
		m.table.SetRows(m.buildRows())
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m libraryModel) handleKey(msg tea.KeyMsg) (libraryModel, tea.Cmd) {
	// Navigation is always available regardless of screen state.
	switch msg.String() {
	case "up", "k", "down", "j":
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		m.table.SetRows(m.buildRows())
		return m, cmd
	}

	switch m.screen {
	case libraryScreenIdle:
		switch msg.String() {
		case "t":
			if len(m.entries) > 0 {
				m.screen = libraryScreenProcessing
				m.entries[m.table.Cursor()].queued = true
				return m.processNext()
			}
		case "e":
			if len(m.entries) > 0 && !m.entries[m.table.Cursor()].enriching {
				e := &m.entries[m.table.Cursor()]
				e.enriching = true
				m.errMsg = ""
				m.table.SetRows(m.buildRows())
				return m, m.enrichEntry(e.rec.Dir, e.rec.Status.RecordedAt.Time)
			}
		case "enter":
			if len(m.entries) > 0 && recording.IsTranscribed(m.entries[m.table.Cursor()].rec.Status.Status) {
				return m, m.openTranscript(m.entries[m.table.Cursor()].rec.Dir)
			}
		case "b":
			return m, func() tea.Msg { return switchToImportMsg{} }
		case "q":
			return m, tea.Quit
		}

	case libraryScreenProcessing:
		switch msg.String() {
		case "t":
			// Queue another recording while one is already running.
			if len(m.entries) > 0 {
				idx := m.table.Cursor()
				e := &m.entries[idx]
				if e.rec.Dir != m.activeDir && !e.queued {
					e.queued = true
					m.table.SetRows(m.buildRows())
				}
			}
		case "e":
			if len(m.entries) > 0 && !m.entries[m.table.Cursor()].enriching {
				e := &m.entries[m.table.Cursor()]
				e.enriching = true
				m.errMsg = ""
				m.table.SetRows(m.buildRows())
				return m, m.enrichEntry(e.rec.Dir, e.rec.Status.RecordedAt.Time)
			}
		case "c":
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
			m.table.SetRows(m.buildRows())
		}
	}
	return m, nil
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

		m.table.SetRows(m.buildRows())
		return m, tea.Batch(workCmd, drainProcessing(ch))
	}

	m.screen = libraryScreenIdle
	m.table.SetRows(m.buildRows())
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

// formatStatusLabel returns a display label for a recording state.
func formatStatusLabel(s recording.State, queued bool) string {
	if queued {
		return "Waiting…"
	}
	switch s {
	case recording.StateCompleted:
		return "Transcribed"
	case recording.StateProcessed:
		return "Processed"
	case recording.StateProcessing:
		return "Processing…"
	case recording.StateDownloaded:
		return "Downloaded"
	case recording.StateEncoded:
		return "Encoded"
	case recording.StateFailed:
		return "Failed"
	case recording.StateInterrupted:
		return "Interrupted"
	case recording.StateNew:
		return "New"
	default:
		return string(s)
	}
}

// enrichEntry looks up the recording's recorded_at time in all ICS files
// found under config.CalendarDir and returns an enrichDoneMsg with the result.
func (m libraryModel) enrichEntry(dir string, recordedAt time.Time) tea.Cmd {
	calDir := m.deps.Config.CalendarDir
	return func() tea.Msg {
		if calDir == "" {
			return enrichDoneMsg{dir: dir, err: fmt.Errorf("calendar_dir not set in config.json")}
		}
		files, err := calendar.FindICSFiles(calDir)
		if err != nil {
			return enrichDoneMsg{dir: dir, err: fmt.Errorf("scan calendar dir: %w", err)}
		}
		event, err := calendar.Match(files, recordedAt)
		if err != nil {
			return enrichDoneMsg{dir: dir, err: err}
		}
		if event == nil {
			return enrichDoneMsg{dir: dir, err: fmt.Errorf("no matching calendar event found")}
		}
		return enrichDoneMsg{
			dir:         dir,
			title:       event.Title,
			attendees:   event.Attendees,
			description: event.Description,
		}
	}
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

	// Indent every line of the table view by two spaces.
	for _, line := range strings.Split(m.table.View(), "\n") {
		b.WriteString("  ")
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n  " + footerDivider + "\n")
	if m.screen == libraryScreenProcessing {
		b.WriteString("  " + formatFooter("t", "queue", "e", "enrich", "c", "cancel") + "\n")
	} else {
		b.WriteString("  " + formatFooter("t", "transcribe", "e", "enrich", "enter", "open") + "\n")
		b.WriteString("  " + footerDivider + "\n")
		b.WriteString("  " + formatFooter("b", "back", "q", "quit") + "\n")
	}

	return b.String()
}
