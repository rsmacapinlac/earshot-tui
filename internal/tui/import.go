package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rsmacapinlac/earshot-tui/internal/config"
	"github.com/rsmacapinlac/earshot-tui/internal/device"
	"github.com/rsmacapinlac/earshot-tui/internal/transcript"
)

type importScreen int

const (
	importScreenScanning importScreen = iota
	importScreenDeviceError
	importScreenReady
	importScreenDownloading
)

type importEntry struct {
	folder   device.Folder
	selected bool
	progress float64
}

type importModel struct {
	deps         Deps
	screen       importScreen
	deviceName   string
	devicePath   string
	entries      []importEntry
	cursor       int
	errMsg       string
	activeFolder string
	downloadCh   <-chan importProgressMsg // live channel; nil when idle
	progressBar  progress.Model
	spinner      spinner.Model
	width        int
	height       int
}

func newImportModel(deps Deps) importModel {
	name, path := config.FirstDevice(deps.Config)
	bar := progress.New(progress.WithDefaultGradient())
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return importModel{
		deps:        deps,
		screen:      importScreenScanning,
		deviceName:  name,
		devicePath:  path,
		progressBar: bar,
		spinner:     sp,
	}
}

func (m importModel) Init() tea.Cmd {
	return tea.Batch(m.scanDevice(), m.spinner.Tick)
}

func (m importModel) scanDevice() tea.Cmd {
	name := m.deviceName
	path := m.devicePath
	recordingsDir := m.deps.RecordingsDir
	return func() tea.Msg {
		if path == "" {
			return deviceScanDoneMsg{err: fmt.Errorf("no device configured")}
		}
		folders, err := device.Scan(path)
		if err != nil {
			return deviceScanDoneMsg{
				err: fmt.Errorf("%s not found at %s. Is it mounted?", name, path),
			}
		}
		var newFolders []device.Folder
		for _, f := range folders {
			if !device.IsImported(f.Name, recordingsDir) {
				newFolders = append(newFolders, f)
			}
		}
		return deviceScanDoneMsg{folders: newFolders}
	}
}

func (m importModel) update(msg tea.Msg) (importModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progressBar.Width = m.width - 20

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		pm, cmd := m.progressBar.Update(msg)
		m.progressBar = pm.(progress.Model)
		return m, cmd

	case deviceScanDoneMsg:
		if msg.err != nil {
			m.screen = importScreenDeviceError
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.screen = importScreenReady
		m.entries = make([]importEntry, len(msg.folders))
		for i, f := range msg.folders {
			m.entries[i] = importEntry{folder: f}
		}

	case importProgressMsg:
		for i := range m.entries {
			if m.entries[i].folder.Name == msg.folderName {
				m.entries[i].progress = msg.fraction
				break
			}
		}
		// Re-queue the channel drain and animate the progress bar.
		return m, tea.Batch(
			m.progressBar.SetPercent(msg.fraction),
			drainImport(m.downloadCh),
		)

	case importFolderDoneMsg:
		m.downloadCh = nil
		if msg.err != nil {
			m.screen = importScreenReady
			m.errMsg = fmt.Sprintf("Download failed: %v", msg.err)
			m.activeFolder = ""
			for i := range m.entries {
				m.entries[i].selected = false
			}
			return m, nil
		}
		// Remove the completed folder from the list.
		var remaining []importEntry
		for _, e := range m.entries {
			if e.folder.Name != msg.folderName {
				remaining = append(remaining, e)
			}
		}
		m.entries = remaining
		if m.cursor >= len(m.entries) && m.cursor > 0 {
			m.cursor--
		}
		m.activeFolder = ""
		return m.downloadNext()

	case deviceDeleteDoneMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("Delete failed: %v", msg.err)
			return m, nil
		}
		var remaining []importEntry
		for _, e := range m.entries {
			if e.folder.Name != msg.folderName {
				remaining = append(remaining, e)
			}
		}
		m.entries = remaining
		if m.cursor >= len(m.entries) && m.cursor > 0 {
			m.cursor--
		}

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m importModel) handleKey(msg tea.KeyMsg) (importModel, tea.Cmd) {
	switch m.screen {
	case importScreenDeviceError:
		switch msg.String() {
		case "r":
			m.screen = importScreenScanning
			return m, tea.Batch(m.scanDevice(), m.spinner.Tick)
		case "l":
			return m, func() tea.Msg { return switchToLibraryMsg{} }
		case "q":
			return m, tea.Quit
		}

	case importScreenReady:
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
				m.entries[m.cursor].selected = !m.entries[m.cursor].selected
			}
		case "i":
			if m.hasSelected() {
				m.screen = importScreenDownloading
				return m.downloadNext()
			}
		case "d":
			if m.hasSelected() {
				var cmds []tea.Cmd
				for _, e := range m.entries {
					if e.selected {
						folder := e.folder
						cmds = append(cmds, func() tea.Msg {
							err := device.Delete(folder)
							return deviceDeleteDoneMsg{folderName: folder.Name, err: err}
						})
					}
				}
				return m, tea.Batch(cmds...)
			}
		case "l":
			return m, func() tea.Msg { return switchToLibraryMsg{} }
		case "q":
			return m, tea.Quit
		}

	case importScreenDownloading:
		if msg.String() == "c" {
			m.screen = importScreenReady
			m.activeFolder = ""
			m.downloadCh = nil
			for i := range m.entries {
				m.entries[i].selected = false
				m.entries[i].progress = 0
			}
		}
	}
	return m, nil
}

func (m importModel) hasSelected() bool {
	for _, e := range m.entries {
		if e.selected {
			return true
		}
	}
	return false
}

// downloadNext starts downloading the next selected folder using a channel-tap
// for live progress, and returns the updated model alongside the initial Cmd.
func (m importModel) downloadNext() (importModel, tea.Cmd) {
	recordingsDir := m.deps.RecordingsDir

	for i := range m.entries {
		if !m.entries[i].selected {
			continue
		}
		folder := m.entries[i].folder
		m.activeFolder = folder.Name
		deviceName := m.deviceName

		ch := make(chan importProgressMsg, 16)
		m.downloadCh = ch

		workCmd := func() tea.Msg {
			localDir, err := device.Download(folder, deviceName, recordingsDir, func(f float64) {
				ch <- importProgressMsg{folderName: folder.Name, fraction: f}
			})
			close(ch)
			return importFolderDoneMsg{folderName: folder.Name, localDir: localDir, err: err}
		}

		return m, tea.Batch(workCmd, drainImport(ch))
	}

	// Nothing left to download.
	m.screen = importScreenReady
	return m, nil
}

func (m importModel) view() string {
	var b strings.Builder

	b.WriteString("\n  " + styleBold.Render("Import Recordings") + "\n\n")

	indicator := ""
	if len(m.deps.Config.DeviceSources) > 1 {
		indicator = " ▾"
	}
	b.WriteString(fmt.Sprintf("  Device: %s%s\n\n", m.deviceName, indicator))

	switch m.screen {
	case importScreenScanning:
		b.WriteString(fmt.Sprintf("  %s Scanning…\n", m.spinner.View()))
		b.WriteString("\n  " + footerDivider + "\n")
		b.WriteString("  " + formatFooter("l", "library", "q", "quit") + "\n")

	case importScreenDeviceError:
		b.WriteString("  " + styleError.Render(m.errMsg) + "\n")
		b.WriteString("\n  " + footerDivider + "\n")
		b.WriteString("  " + formatFooter("r", "retry") + "\n")
		b.WriteString("  " + footerDivider + "\n")
		b.WriteString("  " + formatFooter("l", "library", "q", "quit") + "\n")

	case importScreenReady:
		if len(m.entries) == 0 {
			b.WriteString(fmt.Sprintf("  No new recordings on %s.\n", m.deviceName))
			b.WriteString("\n  " + footerDivider + "\n")
			b.WriteString("  " + formatFooter("l", "library", "q", "quit") + "\n")
			break
		}
		if m.errMsg != "" {
			b.WriteString("  " + styleError.Render(m.errMsg) + "\n\n")
		}
		for i, e := range m.entries {
			cur := "  "
			if i == m.cursor {
				cur = "> "
			}
			checkbox := "[ ]"
			if e.selected {
				checkbox = styleCompleted.Render("[✓]")
			}
			b.WriteString(fmt.Sprintf("%s%s %s\n", cur, checkbox, formatFolderLabel(e.folder)))
		}
		b.WriteString("\n  " + footerDivider + "\n")
		if m.hasSelected() {
			b.WriteString("  " + formatFooter("space", "select", "i", "import", "d", "delete") + "\n")
		} else {
			b.WriteString("  " + formatFooter("space", "select") + "\n")
		}
		b.WriteString("  " + footerDivider + "\n")
		b.WriteString("  " + formatFooter("l", "library", "q", "quit") + "\n")

	case importScreenDownloading:
		for _, e := range m.entries {
			if e.folder.Name == m.activeFolder {
				pct := int(e.progress * 100)
				b.WriteString(fmt.Sprintf("  [↓] %s  Downloading… %s %d%%\n",
					e.folder.Name, m.progressBar.View(), pct))
			} else if e.selected {
				b.WriteString(fmt.Sprintf("  [ ] %s\n", formatFolderLabel(e.folder)))
			}
		}
		b.WriteString("\n  " + footerDivider + "\n")
		b.WriteString("  " + formatFooter("c", "cancel") + "\n")
	}

	return b.String()
}

func formatFolderLabel(f device.Folder) string {
	ts := f.Name
	if !f.Timestamp.IsZero() {
		ts = f.Timestamp.Format("2006-01-02 15:04")
	}
	word := "recordings"
	if f.Count == 1 {
		word = "recording "
	}
	dur := ""
	if f.Duration > 0 {
		dur = "  " + transcript.FormatDuration(f.Duration)
	}
	return fmt.Sprintf("%s  %2d %s%s", ts, f.Count, word, dur)
}
