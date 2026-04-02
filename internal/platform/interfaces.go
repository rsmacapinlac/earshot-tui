package platform

// MountedDevice represents an earshot device found on a mounted volume.
type MountedDevice struct {
	Name      string // human-readable name (hostname from device or derived)
	MountPath string // e.g. /run/media/ritchie/EARSHOT
}

// AppDirs resolves platform-appropriate directories for config, cache, and data.
type AppDirs interface {
	Config() string
	Cache() string
	Data() string
}

// MountScanner discovers earshot devices on mounted volumes.
type MountScanner interface {
	Scan() ([]MountedDevice, error)
}

// AudioPlayer plays an audio file using a system tool.
type AudioPlayer interface {
	Play(filePath string) error
	Stop() error
}

// PythonResolver locates a suitable Python executable on the host.
type PythonResolver interface {
	Find() (path string, version string, err error)
	VenvPython(venvDir string) string
}
