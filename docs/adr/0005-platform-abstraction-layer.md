# ADR-0005: Platform Abstraction Layer

**Status:** Accepted

## Context

earshot-tui v1 targets macOS and Linux. Windows support is deferred but must
not require structural changes to the codebase when it is added (NFR-2).

Several operations behave differently across platforms:

| Operation | macOS | Linux | Windows |
|---|---|---|---|
| Find Python executable | `python3` in PATH | `python3` in PATH | `python`, `py` launcher, or Microsoft Store path |
| venv Python path | `venv/bin/python` | `venv/bin/python` | `venv\Scripts\python.exe` |
| USB mount paths | `/Volumes/` | `/media/$USER/`, `/mnt/` | Drive letters via `GetLogicalDrives` |
| Audio playback | `afplay` (built-in) | `ffplay` (requires install) | `ffplay` or PowerShell |
| App config dir | `~/Library/Application Support/` | `~/.config/` | `%APPDATA%` |
| App cache dir | `~/Library/Caches/` | `~/.cache/` | `%LOCALAPPDATA%` |

Without abstraction, platform differences scatter through the codebase and
Windows support becomes a risky, wide-surface refactor.

## Decision

Define four platform interfaces in Go from the outset. Each has a macOS/Linux
implementation at launch. Windows implementations are added later without
touching any other code.

```go
// PythonResolver locates a suitable Python executable on the host.
type PythonResolver interface {
    Find() (path string, version string, err error)
    VenvPython(venvDir string) string
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

// AppDirs resolves platform-appropriate directories for config, cache, data.
type AppDirs interface {
    Config() string
    Cache() string
    Data() string
}
```

Concrete implementations live in platform-tagged files:

```
internal/platform/
  python_unix.go       // build tag: !windows
  python_windows.go    // build tag: windows
  mount_darwin.go      // build tag: darwin
  mount_linux.go       // build tag: linux
  mount_windows.go     // build tag: windows
  audio_darwin.go      // build tag: darwin
  audio_linux.go       // build tag: linux
  audio_windows.go     // build tag: windows
  dirs.go              // uses os.UserConfigDir() — already cross-platform
```

## Consequences

- All platform-specific behaviour is contained in `internal/platform/`.
  No other package imports OS detection logic.
- Adding Windows support is a matter of implementing the four interfaces in
  `*_windows.go` files — no changes to TUI, database, or processing code.
- `MountScanner` on Linux must handle both `/media/$USER/` and `/mnt/` and
  filter for directories containing an `earshot.db` file.
- `AudioPlayer` on Linux depends on `ffplay` being installed. If not found,
  the error message tells the user how to install it (../ux-standards.md §6).
  On macOS, `afplay` is always available.
- `PythonResolver` must handle the case where Python is not installed and
  return a clear error (not a panic) — the app surfaces this on first run
  with platform-specific install instructions.
- Go's `os.UserConfigDir()` and `os.UserCacheDir()` already handle macOS,
  Linux, and Windows correctly — `AppDirs` wraps these with earshot-tui
  subdirectory naming for convenience.
