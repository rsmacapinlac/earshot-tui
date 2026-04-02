//go:build linux

package platform

import (
	"os"
	"path/filepath"
)

type linuxMountScanner struct{}

// NewMountScanner returns a MountScanner for Linux (/media/$USER and /mnt).
func NewMountScanner() MountScanner {
	return &linuxMountScanner{}
}

func (s *linuxMountScanner) Scan() ([]MountedDevice, error) {
	user := os.Getenv("USER")
	roots := []string{
		filepath.Join("/media", user),
		"/run/media/" + user,
		"/mnt",
	}
	return scanRoots(roots)
}
