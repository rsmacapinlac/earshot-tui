package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rsmacapinlac/earshot-tui/internal/config"
	"github.com/rsmacapinlac/earshot-tui/internal/device"
	"github.com/rsmacapinlac/earshot-tui/internal/processor"
	"github.com/rsmacapinlac/earshot-tui/internal/recording"
)

// ---- navigation -------------------------------------------------------------

// setupCompleteMsg is sent when the setup wizard successfully writes config.json.
type setupCompleteMsg struct {
	cfg *config.Config
}

// switchToImportMsg navigates the root to the import screen.
type switchToImportMsg struct{}

// switchToLibraryMsg navigates the root to the library screen.
type switchToLibraryMsg struct{}

// ---- preflight --------------------------------------------------------------

// preflightCheckMsg reports the result of one preflight check.
type preflightCheckMsg struct {
	index int   // which check (0, 1, 2)
	err   error // nil = passed
}

// preflightDoneMsg is sent when all preflight checks pass.
type preflightDoneMsg struct{}

// ---- device scan / import ---------------------------------------------------

// deviceScanDoneMsg carries the result of scanning a device for new folders.
type deviceScanDoneMsg struct {
	folders []device.Folder
	err     error
}

// importProgressMsg reports download progress for a single folder.
type importProgressMsg struct {
	folderName string
	fraction   float64
}

// importFolderDoneMsg is sent when one folder finishes downloading.
type importFolderDoneMsg struct {
	folderName string
	localDir   string
	err        error
}

// deviceDeleteDoneMsg is sent when a folder has been deleted from the device.
type deviceDeleteDoneMsg struct {
	folderName string
	err        error
}

// ---- library / processing ---------------------------------------------------

// libraryLoadedMsg carries the list of local recordings for the library screen.
type libraryLoadedMsg struct {
	recordings []recording.Recording
	err        error
}

// processingProgressMsg reports transcription progress for a recording.
type processingProgressMsg struct {
	dir      string
	progress processor.Progress
}

// processingDoneMsg is sent when one recording finishes (or fails) processing.
type processingDoneMsg struct {
	dir string
	err error
}

// enrichDoneMsg is sent when calendar enrichment finishes for a recording.
type enrichDoneMsg struct {
	dir         string
	title       string
	attendees   []string
	description string
	err         error // non-nil on failure or no match
}

// ---- channel-tap helpers ----------------------------------------------------
// These allow a background goroutine to stream tea.Msg values one at a time
// through a channel without blocking the Bubble Tea event loop.
//
// Usage pattern:
//   ch := make(chan importProgressMsg, 16)
//   go func() { /* send to ch */ ; close(ch) }()
//   return tea.Batch(workCmd, drainImport(ch))
//
// In update(), when importProgressMsg arrives:
//   return m, tea.Batch(otherCmds..., drainImport(m.downloadCh))

// drainImport returns a Cmd that reads one importProgressMsg from ch.
// Returns nil (ignored by Bubble Tea) when the channel is closed.
func drainImport(ch <-chan importProgressMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

// drainProcessing returns a Cmd that reads one processingProgressMsg from ch.
// Returns nil when the channel is closed.
func drainProcessing(ch <-chan processingProgressMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}
