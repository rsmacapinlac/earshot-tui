// Package calendar finds and matches ICS calendar events to recording times.
package calendar

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// matchWindow is the maximum delta between a recording and an event start.
const matchWindow = 10 * time.Minute

// Event holds the fields extracted from a VEVENT component.
type Event struct {
	Title       string
	Description string
	Attendees   []string
	Start       time.Time
	End         time.Time // zero if absent; defaultDuration is assumed in that case
}

// defaultDuration is assumed when an event has no DTEND or DURATION.
const defaultDuration = 60 * time.Minute

// FindICSFiles recursively finds all .ics files under dir.
// Unreadable subdirectories are silently skipped.
func FindICSFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".ics" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// Match searches files for the best-matching calendar event for time t.
// A match is valid if t falls within [event.Start - matchWindow, event.End + matchWindow].
// Among valid matches, the one whose start is closest to t wins.
// Returns nil if no match is found.
func Match(files []string, t time.Time) (*Event, error) {
	var best *Event
	bestDelta := time.Duration(1<<63 - 1)

	for _, f := range files {
		events, err := parseFile(f)
		if err != nil {
			continue // skip unreadable or malformed files
		}
		for i := range events {
			ev := &events[i]
			end := ev.End
			if end.IsZero() {
				end = ev.Start.Add(defaultDuration)
			}
			lo := ev.Start.Add(-matchWindow)
			hi := end.Add(matchWindow)
			if t.Before(lo) || t.After(hi) {
				continue
			}
			delta := t.Sub(ev.Start)
			if delta < 0 {
				delta = -delta
			}
			if delta < bestDelta {
				bestDelta = delta
				cp := *ev
				best = &cp
			}
		}
	}
	return best, nil
}

// ---- parsing ----------------------------------------------------------------

func parseFile(path string) ([]Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseICS(f)
}

func parseICS(r io.Reader) ([]Event, error) {
	lines := unfoldLines(r)

	var events []Event
	var cur *Event
	var cancelled bool

	for _, line := range lines {
		name, params, value := parseLine(line)

		switch name {
		case "BEGIN":
			if value == "VEVENT" {
				cur = &Event{}
				cancelled = false
			}
		case "END":
			if value == "VEVENT" && cur != nil {
				if !cancelled && !cur.Start.IsZero() {
					events = append(events, *cur)
				}
				cur = nil
			}
		case "SUMMARY":
			if cur != nil {
				cur.Title = unescapeICS(value)
			}
		case "DESCRIPTION":
			if cur != nil {
				cur.Description = unescapeICS(value)
			}
		case "STATUS":
			if cur != nil && strings.ToUpper(value) == "CANCELLED" {
				cancelled = true
			}
		case "DTSTART":
			if cur != nil {
				tzid := extractParam(params, "TZID")
				t, err := parseDTSTART(value, tzid)
				if err == nil {
					cur.Start = t
				}
			}
		case "DTEND":
			if cur != nil {
				tzid := extractParam(params, "TZID")
				t, err := parseDTSTART(value, tzid) // same format as DTSTART
				if err == nil {
					cur.End = t
				}
			}
		case "DURATION":
			if cur != nil && !cur.Start.IsZero() {
				if d, err := parseICALDuration(value); err == nil {
					cur.End = cur.Start.Add(d)
				}
			}
		case "ATTENDEE":
			if cur != nil {
				cn := extractParam(params, "CN")
				if cn != "" {
					cur.Attendees = append(cur.Attendees, cn)
				}
			}
		}
	}

	return events, nil
}

// unfoldLines reads an ICS stream and joins continuation lines (RFC 5545 §3.1).
func unfoldLines(r io.Reader) []string {
	sc := bufio.NewScanner(r)
	var lines []string
	var cur strings.Builder

	for sc.Scan() {
		line := sc.Text()
		// Strip trailing CR if present (CRLF line endings).
		line = strings.TrimRight(line, "\r")
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			cur.WriteString(line[1:]) // continuation
		} else {
			if cur.Len() > 0 {
				lines = append(lines, cur.String())
				cur.Reset()
			}
			cur.WriteString(line)
		}
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}

// parseLine splits "NAME;PARAMS:VALUE" into its three parts.
func parseLine(line string) (name, params, value string) {
	colon := strings.IndexByte(line, ':')
	if colon < 0 {
		return strings.ToUpper(line), "", ""
	}
	nameParams := line[:colon]
	value = line[colon+1:]
	semi := strings.IndexByte(nameParams, ';')
	if semi < 0 {
		return strings.ToUpper(nameParams), "", value
	}
	return strings.ToUpper(nameParams[:semi]), nameParams[semi+1:], value
}

// extractParam finds a named parameter in a semicolon-separated params string.
// e.g. extractParam("TZID=America/Vancouver;CN=Ritchie", "CN") → "Ritchie"
func extractParam(params, key string) string {
	prefix := strings.ToUpper(key) + "="
	for _, p := range strings.Split(params, ";") {
		if strings.HasPrefix(strings.ToUpper(p), prefix) {
			val := p[len(prefix):]
			// Strip surrounding quotes.
			if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
				val = val[1 : len(val)-1]
			}
			return val
		}
	}
	return ""
}

// parseDTSTART parses a DTSTART value with an optional TZID parameter.
// All-day events (DATE values, 8 chars) are rejected.
func parseDTSTART(value, tzid string) (time.Time, error) {
	if len(value) == 8 {
		return time.Time{}, fmt.Errorf("all-day event")
	}
	if strings.HasSuffix(value, "Z") {
		return time.Parse("20060102T150405Z", value)
	}
	loc := time.Local
	if tzid != "" {
		if l, err := time.LoadLocation(tzid); err == nil {
			loc = l
		}
	}
	return time.ParseInLocation("20060102T150405", value, loc)
}

// parseICALDuration parses an iCalendar DURATION value (RFC 5545 §3.3.6).
// Supports the common subset: P[nW][nD][T[nH][nM][nS]]
func parseICALDuration(s string) (time.Duration, error) {
	if len(s) == 0 || s[0] != 'P' {
		return 0, fmt.Errorf("invalid DURATION: %q", s)
	}
	s = s[1:]
	var total time.Duration
	inTime := false
	for len(s) > 0 {
		if s[0] == 'T' {
			inTime = true
			s = s[1:]
			continue
		}
		i := 0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		if i == 0 || i >= len(s) {
			break
		}
		var n int
		fmt.Sscanf(s[:i], "%d", &n)
		unit := s[i]
		s = s[i+1:]
		switch unit {
		case 'W':
			total += time.Duration(n) * 7 * 24 * time.Hour
		case 'D':
			total += time.Duration(n) * 24 * time.Hour
		case 'H':
			if inTime {
				total += time.Duration(n) * time.Hour
			}
		case 'M':
			if inTime {
				total += time.Duration(n) * time.Minute
			}
		case 'S':
			if inTime {
				total += time.Duration(n) * time.Second
			}
		}
	}
	return total, nil
}

// unescapeICS decodes ICS text escapes.
func unescapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\N", "\n")
	s = strings.ReplaceAll(s, "\\,", ",")
	s = strings.ReplaceAll(s, "\\;", ";")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return strings.TrimSpace(s)
}
