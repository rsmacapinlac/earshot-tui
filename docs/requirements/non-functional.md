# Non-Functional Requirements

## Platform

- NFR-1: v1 targets macOS and Linux.
- NFR-2: Windows support is deferred to a future version, but architectural
  decisions must not foreclose it. Platform-specific paths (mount points,
  audio playback commands, filesystem conventions) must be isolated behind
  an abstraction layer so Windows support can be added without structural
  changes.

## Performance

- NFR-3: The app must remain responsive during processing. Transcription and
  diarization must run in background workers; the TUI must not freeze or
  become unnavigable.
- NFR-4: Scanning a device for new recordings must complete within 5 seconds
  for up to 500 recordings.
- NFR-5: On subsequent launches, app launch to connect screen must take
  under 3 seconds. First launch is exempt — it involves venv creation,
  dependency installation, and model downloads, which must show a clear
  progress indicator but have no time constraint.

## Storage

- NFR-6: The local application database (SQLite) is the source of truth for
  recording state on the user's computer.
- NFR-7: The app must never corrupt or unintentionally modify the earshot
  device's SQLite database. Writes to the device database are limited to
  explicit user-initiated delete actions.

## Reliability

- NFR-8: If the device is disconnected mid-session, the app handles this
  gracefully — no crash, clear error, return to connect screen.
- NFR-9: The app is safe to close at any time. State is persisted to the
  local database before any significant operation. In-progress processing
  interrupted by a close is recoverable on next launch.

## Privacy

- NFR-10: No audio, transcript, or metadata leaves the user's computer.
  All processing is strictly local. No telemetry, no analytics.
