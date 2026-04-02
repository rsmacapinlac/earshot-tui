//go:build darwin

package platform

type darwinMountScanner struct{}

// NewMountScanner returns a MountScanner for macOS (/Volumes).
func NewMountScanner() MountScanner {
	return &darwinMountScanner{}
}

func (s *darwinMountScanner) Scan() ([]MountedDevice, error) {
	return scanRoots([]string{"/Volumes"})
}
