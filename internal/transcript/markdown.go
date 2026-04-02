// Package transcript renders processor output as Markdown transcripts.
package transcript

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rsmacapinlac/earshot-tui/internal/processor"
	"github.com/rsmacapinlac/earshot-tui/internal/recording"
)

// Write renders result as a Markdown transcript and saves it to
// recording.TranscriptPath(dir). The write is not atomic because transcripts
// are generated once and not updated — a partial file on crash is benign
// (the recording stays in "processing" / "interrupted" state).
func Write(dir string, status *recording.Status, result *processor.Result) error {
	var b strings.Builder

	fmt.Fprintf(&b, "# Recording — %s\n\n",
		status.RecordedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "**Device:** %s\n", status.Device)
	fmt.Fprintf(&b, "**Duration:** %s\n", FormatDuration(result.Duration))
	fmt.Fprintf(&b, "**Processed:** %s\n", time.Now().UTC().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "\n---\n\n")

	for _, seg := range result.Segments {
		text := strings.TrimSpace(seg.Text)
		if text == "" {
			continue
		}
		fmt.Fprintf(&b, "%s %s\n\n", formatTimestamp(seg.Start, result.Duration), text)
	}

	return os.WriteFile(recording.TranscriptPath(dir), []byte(b.String()), 0o644)
}

// formatTimestamp returns [MM:SS] for recordings under one hour, [HH:MM:SS] otherwise.
func formatTimestamp(secs, totalDuration float64) string {
	s := int(secs)
	if totalDuration >= 3600 {
		h := s / 3600
		m := (s % 3600) / 60
		sec := s % 60
		return fmt.Sprintf("[%02d:%02d:%02d]", h, m, sec)
	}
	m := s / 60
	sec := s % 60
	return fmt.Sprintf("[%02d:%02d]", m, sec)
}

// FormatDuration returns a human-readable duration string like "3m 42s".
func FormatDuration(secs float64) string {
	total := int(secs)
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	switch {
	case h > 0:
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	case m > 0:
		return fmt.Sprintf("%dm %ds", m, s)
	default:
		return fmt.Sprintf("%ds", s)
	}
}
