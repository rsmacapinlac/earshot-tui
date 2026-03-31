# ADR-0001: Go as Implementation Language

**Status:** Accepted

## Context

earshot-tui needs to manage USB-mounted devices, maintain local state, spawn
and manage a background processing subprocess, play audio via system tools,
and present a responsive terminal UI — all while keeping distribution simple
for the end user.

The initial assumption was Python (matching the earshot device codebase), but
the processing stack was resolved separately (see ADR-0003). With processing
decoupled, the TUI language choice is free to optimise for distribution,
concurrency, and UX.

Candidates considered:

- **Python**: Native access to processing libraries, but GIL complicates
  background processing alongside a live TUI, and distribution requires
  a Python runtime on the host.
- **Rust**: Strong performance and safety guarantees, but highest
  implementation effort and smallest ecosystem for TUI tooling in this domain.
- **Go**: Compiles to a single native binary, goroutines handle background
  processing cleanly, strong cross-platform support, and the Charm ecosystem
  provides first-class TUI tooling.

## Decision

Use **Go (1.22+)** as the implementation language for earshot-tui.

## Consequences

- Distribution is a single native binary with no runtime dependency on Go.
- Goroutines provide a natural model for running transcription in the
  background while keeping the TUI responsive (PROC-14, NFR-3).
- Go's `//go:embed` directive enables bundling the Python processor script
  and its dependency manifest directly into the binary (see ADR-0003).
- `os.UserConfigDir()`, `os.UserCacheDir()` resolve platform-appropriate
  paths on macOS, Linux, and Windows without custom logic.
- Cross-compilation is straightforward: a single build pipeline produces
  binaries for macOS (arm64, amd64), Linux (amd64, arm64), and eventually
  Windows (amd64).
- The Go codebase is separate from the earshot Python codebase — contributors
  need Go familiarity, not Python, to work on the TUI.
