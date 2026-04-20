// Package recording manages per-recording state stored in status.json files.
package recording

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FlexTime is a time.Time that also accepts ISO 8601 timestamps without a
// timezone (written by earshot Pi), treating them as local time.
type FlexTime struct{ time.Time }

var flexFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05",
}

func (ft *FlexTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" {
		return nil
	}
	for _, layout := range flexFormats {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			ft.Time = t
			return nil
		}
	}
	return fmt.Errorf("recording: cannot parse time %q", s)
}

func (ft FlexTime) MarshalJSON() ([]byte, error) {
	return ft.Time.MarshalJSON()
}

// State represents the lifecycle state of a recording.
type State string

const (
	StateNew         State = "new"
	StateDownloaded  State = "downloaded"
	StateEncoded     State = "encoded" // set by earshot Pi after opus encoding; equivalent to downloaded
	StateProcessing  State = "processing"
	StateCompleted   State = "transcribed"
	StateProcessed   State = "processed" // set by external tools after notes generation; treated as completed
	StateFailed      State = "failed"
	StateInterrupted State = "interrupted"
)

// Status is the data persisted in each recording's status.json.
type Status struct {
	Status        State     `json:"status"`
	Device        string    `json:"device"`
	RecordedAt    FlexTime  `json:"recorded_at"`
	Duration      float64   `json:"duration"` // seconds
	DownloadedAt  *FlexTime `json:"downloaded_at,omitempty"`
	TranscribedAt *FlexTime `json:"transcribed_at,omitempty"`
	Error         string    `json:"error,omitempty"`

	// Enrichment fields — populated by calendar matching.
	Title       string    `json:"title,omitempty"`
	Attendees   []string  `json:"attendees,omitempty"`
	Description string    `json:"description,omitempty"`
	EnrichedAt  *FlexTime `json:"enriched_at,omitempty"`
}

// Recording pairs a folder path with its persisted status.
type Recording struct {
	Dir    string // absolute path to the recording folder
	Status Status
}

// LoadStatus reads and parses status.json from dir.
func LoadStatus(dir string) (*Status, error) {
	path := filepath.Join(dir, "status.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var s Status
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &s, nil
}

// SaveStatus atomically writes s to dir/status.json.
func SaveStatus(dir string, s *Status) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal status: %w", err)
	}
	dest := filepath.Join(dir, "status.json")
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename %s → %s: %w", tmp, dest, err)
	}
	return nil
}

// List scans recordingsDir and returns all valid recordings sorted most-recent-first.
// Folders without a readable status.json are silently skipped.
func List(recordingsDir string) ([]Recording, error) {
	entries, err := os.ReadDir(recordingsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read recordings dir %s: %w", recordingsDir, err)
	}
	var recs []Recording
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(recordingsDir, e.Name())
		s, err := LoadStatus(dir)
		if err != nil {
			continue
		}
		recs = append(recs, Recording{Dir: dir, Status: *s})
	}
	sort.Slice(recs, func(i, j int) bool {
		return recs[i].Status.RecordedAt.After(recs[j].Status.RecordedAt.Time)
	})
	return recs, nil
}

// RecoverInterrupted resets all recordings stuck in "processing" to "interrupted".
// Called on startup to recover from a crash or forced quit.
func RecoverInterrupted(recordingsDir string) error {
	entries, err := os.ReadDir(recordingsDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", recordingsDir, err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(recordingsDir, e.Name())
		s, err := LoadStatus(dir)
		if err != nil || s.Status != StateProcessing {
			continue
		}
		s.Status = StateInterrupted
		if err := SaveStatus(dir, s); err != nil {
			return fmt.Errorf("recover %s: %w", dir, err)
		}
	}
	return nil
}

// FolderName returns the canonical local folder name for a recording timestamp.
func FolderName(t time.Time) string {
	return t.UTC().Format("2006-01-02_15-04-05")
}

// AudioPath returns the path to the recording's .opus file.
func AudioPath(dir string) string {
	return filepath.Join(dir, "recording.opus")
}

// TranscriptPath returns the path to the recording's Markdown transcript.
func TranscriptPath(dir string) string {
	return filepath.Join(dir, "transcript.md")
}

// StatusPath returns the path to the recording's status.json file.
func StatusPath(dir string) string {
	return filepath.Join(dir, "status.json")
}

// IsTranscribed reports whether a recording has been transcribed or fully processed.
func IsTranscribed(s State) bool {
	return s == StateCompleted || s == StateProcessed
}

// IsRetryable reports whether a recording in this state can be re-processed.
func IsRetryable(s State) bool {
	return s == StateFailed || s == StateInterrupted || s == StateDownloaded || s == StateEncoded
}
