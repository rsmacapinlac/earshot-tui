package platform

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// recordingDirPattern matches the earshot folder naming format: 20260401T083114
var recordingDirPattern = regexp.MustCompile(`^\d{8}T\d{6}$`)

// IsEarshotDevice reports whether path is a mounted earshot device.
// Detection: the directory contains at least one timestamp-format folder
// (e.g. 20260401T083114) that holds at least one .opus file.
func IsEarshotDevice(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() || !recordingDirPattern.MatchString(e.Name()) {
			continue
		}
		sub, err := os.ReadDir(filepath.Join(path, e.Name()))
		if err != nil {
			continue
		}
		for _, f := range sub {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".opus" {
				return true
			}
		}
	}
	return false
}

// DeviceHostname reads the preferred_hostname file from the device filesystem,
// falling back to the mount directory's base name.
func DeviceHostname(mountPath string) string {
	data, err := os.ReadFile(filepath.Join(mountPath, "preferred_hostname"))
	if err == nil {
		if name := strings.TrimSpace(string(data)); name != "" {
			return name
		}
	}
	return filepath.Base(mountPath)
}

// scanRoots walks each root directory looking for earshot devices.
func scanRoots(roots []string) ([]MountedDevice, error) {
	var devices []MountedDevice
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			mountPath := filepath.Join(root, entry.Name())
			if IsEarshotDevice(mountPath) {
				devices = append(devices, MountedDevice{
					Name:      DeviceHostname(mountPath),
					MountPath: mountPath,
				})
			}
		}
	}
	return devices, nil
}
