// Package processor manages the embedded Python venv and runs processor.py.
package processor

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Progress is a processing update parsed from processor.py stderr.
type Progress struct {
	Phase    string
	Fraction float64 // 0.0–1.0; -1 when not applicable (e.g. "complete")
}

// Result is the JSON payload written by processor.py to stdout on success.
type Result struct {
	Version  string    `json:"version"`
	Duration float64   `json:"duration"`
	Segments []Segment `json:"segments"`
}

// Segment is one transcription segment from the Result.
type Segment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// Manager handles the Python venv and processor.py lifecycle.
type Manager struct {
	pythonPath   string // system Python 3.10+ executable
	processorDir string // ~config/processor/ — extracted processor.py lives here
	venvDir      string // ~config/venv/
	cacheDir     string // ~cache/ — HuggingFace models downloaded here
}

// NewManager creates a Manager. pythonPath may be empty only when
// EARSHOT_PROCESSOR_STUB=1 is set (stub mode bypasses Python entirely).
func NewManager(pythonPath, configDir, cacheDir string) *Manager {
	return &Manager{
		pythonPath:   pythonPath,
		processorDir: filepath.Join(configDir, "processor"),
		venvDir:      filepath.Join(configDir, "venv"),
		cacheDir:     cacheDir,
	}
}

// Setup extracts processor files and ensures the Python venv is ready.
// progressFn receives human-readable status strings for display in the TUI.
// In stub mode (EARSHOT_PROCESSOR_STUB=1) this is a no-op.
func (m *Manager) Setup(progressFn func(string)) error {
	if os.Getenv("EARSHOT_PROCESSOR_STUB") == "1" {
		return nil
	}

	progressFn("Extracting processor files…")
	if err := m.extractFiles(); err != nil {
		return fmt.Errorf("extract processor files: %w", err)
	}

	venvPython := m.venvPython()
	if _, err := os.Stat(venvPython); os.IsNotExist(err) {
		progressFn("Creating Python virtual environment…")
		cmd := exec.Command(m.pythonPath, "-m", "venv", m.venvDir)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("create venv: %w\n%s", err, out)
		}
	}

	if m.requirementsChanged() {
		progressFn("Installing Python dependencies (first run may take a moment)…")
		pip := filepath.Join(m.venvDir, "bin", "pip")
		req := filepath.Join(m.processorDir, "requirements.txt")
		cmd := exec.Command(pip, "install", "--quiet", "-r", req)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("pip install: %w\n%s", err, out)
		}
		if err := m.saveRequirementsHash(); err != nil {
			return err
		}
	}

	return nil
}

// Run transcribes audioPath, streaming Progress updates via progressFn.
// Returns the parsed Result or an error if the processor exits non-zero.
func (m *Manager) Run(audioPath string, progressFn func(Progress)) (*Result, error) {
	if os.Getenv("EARSHOT_PROCESSOR_STUB") == "1" {
		return m.runStub(progressFn)
	}

	script := filepath.Join(m.processorDir, "processor.py")
	cmd := exec.Command(m.venvPython(), script, audioPath)
	cmd.Env = append(os.Environ(),
		"EARSHOT_CACHE_DIR="+filepath.Join(m.cacheDir, "huggingface"),
		"PYTHONUNBUFFERED=1",
	)

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start processor: %w", err)
	}

	// Parse stderr progress lines in a goroutine.
	go func() {
		sc := bufio.NewScanner(stderrPipe)
		for sc.Scan() {
			if p, ok := parseProgress(sc.Text()); ok {
				progressFn(p)
			}
		}
	}()

	stdoutData, err := io.ReadAll(stdoutPipe)
	if err != nil {
		return nil, fmt.Errorf("read stdout: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("processor exited with error: %w", err)
	}

	var result Result
	if err := json.Unmarshal(stdoutData, &result); err != nil {
		return nil, fmt.Errorf("parse processor output: %w", err)
	}
	return &result, nil
}

// ---- internal ---------------------------------------------------------------

func (m *Manager) venvPython() string {
	return filepath.Join(m.venvDir, "bin", "python")
}

func (m *Manager) extractFiles() error {
	if err := os.MkdirAll(m.processorDir, 0o755); err != nil {
		return err
	}
	files := map[string][]byte{
		"processor.py":     processorPy,
		"requirements.txt": requirementsTxt,
	}
	for name, data := range files {
		dest := filepath.Join(m.processorDir, name)
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}

func (m *Manager) requirementsHash() string {
	h := sha256.Sum256(requirementsTxt)
	return hex.EncodeToString(h[:])
}

func (m *Manager) hashFile() string {
	return filepath.Join(m.processorDir, ".requirements.sha256")
}

func (m *Manager) requirementsChanged() bool {
	data, err := os.ReadFile(m.hashFile())
	if err != nil {
		return true
	}
	return strings.TrimSpace(string(data)) != m.requirementsHash()
}

func (m *Manager) saveRequirementsHash() error {
	return os.WriteFile(m.hashFile(), []byte(m.requirementsHash()), 0o644)
}

func (m *Manager) runStub(_ func(Progress)) (*Result, error) {
	return &Result{
		Version:  "1",
		Duration: 10.0,
		Segments: []Segment{
			{Start: 0.5, End: 3.0, Text: "This is a stub transcription."},
			{Start: 3.5, End: 6.0, Text: "The processor is running in stub mode."},
		},
	}, nil
}

func parseProgress(line string) (Progress, bool) {
	if !strings.HasPrefix(line, "PROGRESS:") {
		return Progress{}, false
	}
	rest := strings.TrimPrefix(line, "PROGRESS:")
	parts := strings.SplitN(rest, ":", 2)
	p := Progress{Phase: parts[0], Fraction: -1}
	if len(parts) == 2 {
		fmt.Sscanf(parts[1], "%f", &p.Fraction)
	}
	return p, true
}
