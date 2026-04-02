//go:build !windows

package platform

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type unixPythonResolver struct{}

// NewPythonResolver returns a PythonResolver for Unix-like systems.
func NewPythonResolver() PythonResolver {
	return &unixPythonResolver{}
}

// Find locates a Python 3.10+ executable on PATH.
func (r *unixPythonResolver) Find() (string, string, error) {
	candidates := []string{"python3", "python"}
	for _, name := range candidates {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		version, err := pythonVersion(path)
		if err != nil {
			continue
		}
		if isSupportedVersion(version) {
			return path, version, nil
		}
	}
	return "", "", fmt.Errorf("Python 3.10+ not found on PATH; " +
		"install it with your system package manager (e.g. sudo apt install python3)")
}

// VenvPython returns the path to the python binary inside a venv on Unix.
func (r *unixPythonResolver) VenvPython(venvDir string) string {
	return filepath.Join(venvDir, "bin", "python")
}

func pythonVersion(path string) (string, error) {
	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		// python2 writes to stderr; try combined
		out, err = exec.Command(path, "--version").CombinedOutput()
		if err != nil {
			return "", err
		}
	}
	// Output: "Python 3.11.4\n"
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) < 2 {
		return "", fmt.Errorf("unexpected python --version output: %q", out)
	}
	return fields[1], nil
}

func isSupportedVersion(version string) bool {
	var major, minor int
	if _, err := fmt.Sscanf(version, "%d.%d", &major, &minor); err != nil {
		return false
	}
	return major == 3 && minor >= 10
}
