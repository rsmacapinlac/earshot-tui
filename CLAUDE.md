# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run from source
go run ./cmd/earshot-tui

# Run all tests
go test ./... -count=1

# Run a single package's tests
go test ./internal/db/... -count=1

# Check formatting (must produce no output)
test -z "$(gofmt -l .)"

# Static analysis
go vet ./...
staticcheck ./...   # version 2024.1.1
```

Set `EARSHOT_PROCESSOR_STUB=1` to bypass Whisper/pyannote and return fake output — useful for exercising TUI and DB flows without running ML inference.

System dependencies required: **ffmpeg** on `$PATH`, **Python 3.10–3.12**.

## Architecture

earshot-tui connects to a Raspberry Pi earshot USB device, downloads audio recordings, and produces speaker-labeled Markdown transcripts using locally-run ML models (Whisper + pyannote).

### Entry point and wiring

`cmd/earshot-tui/main.go` initialises everything: `AppDirs` → SQLite DB (migrations + interrupt recovery) → `PythonResolver` → `processor.Manager` → all platform interfaces → `tui.Deps` bundle → Bubble Tea program.

### TUI state machine (`internal/tui/`)

`Root` is the top-level Bubble Tea model. On construction it inspects DB config and the venv to decide whether to start in the **setup wizard** (`setupModel`) or the **main app** (`appModel`). A `setupCompleteMsg` transitions between them.

The main app flow is a linear screen progression:

```
connect → device selection → disposition → processing → library
```

Each screen is its own Bubble Tea model. Async work (scanning, downloading, processing) is dispatched as `tea.Cmd` and results returned as custom message types.

### Platform abstraction (`internal/platform/`)

Four interfaces defined in `interfaces.go` (ADR-0005) isolate OS-specific behaviour:

- `AppDirs` — XDG/macOS standard dirs (`~/.config`, `~/.cache`, `~/.local/share` on Linux)
- `MountScanner` — scans `/media/*` (Linux) or `/Volumes/*` (macOS) for earshot devices
- `AudioPlayer` — plays audio via system tools; separate `*_linux.go` / `*_darwin.go` impls
- `PythonResolver` — locates Python 3.10+ on the host

Platform-specific files follow the `*_darwin.go` / `*_linux.go` naming convention and are selected at compile time via build tags.

### Processor (`internal/processor/`)

`processor.Manager` manages a Python venv under `~/.config/earshot-tui/venv/`. The embedded `processor.py` is extracted to `~/.config/earshot-tui/processor/` on first run, then executed as a subprocess. It writes `PROGRESS <phase> <fraction>` lines to stderr and a JSON result to stdout. Progress is parsed in Go and forwarded into the Bubble Tea event loop.

### Database (`internal/db/`)

SQLite via `modernc.org/sqlite` (pure-Go, no CGo). Three tables: `devices`, `recordings`, `config`. Schema changes go through numbered migrations in `db.go`. `RecoverInterrupted` resets any recordings stuck in `processing` state back to `downloaded` at startup.

Recording lifecycle: `new → downloaded → (processing → processed | failed | interrupted) | skipped`.

### Transcript generation (`internal/transcript/`)

`markdown.go` renders the JSON processor output into Markdown. `rewrite.go` replaces generic speaker labels (e.g. `SPEAKER_00`) with user-assigned names in an existing transcript file.

### Data layout

| Directory | Contents |
|-----------|----------|
| `~/.config/earshot-tui/` | `processor/` (extracted Python scripts), `venv/` |
| `~/.cache/earshot-tui/` | HuggingFace model cache, PyTorch cache |
| `~/.local/share/earshot-tui/` | `earshot-tui.db`, `audio/`, `transcripts/` |
