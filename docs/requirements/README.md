# Earshot TUI — Requirements

A terminal user interface that runs on a desktop/laptop to download, process,
and manage recordings from one or more [earshot](https://github.com/rsmacapinlac/earshot)
devices. Transcription and speaker diarization run entirely locally.

## Scope (v1)

- USB-connected earshot device management (auto-detect, hostname-based naming)
- Local transcription and diarization
- Markdown transcript output with timestamps
- Per-recording disposition (download, skip, delete from device)
- Inline audio playback for speaker identification
- Optional per-recording speaker renaming
- No network calls, no authentication, no API integration

## Out of Scope (v1)

- Drop folder / watch folder support
- Speaker profiles persisted across recordings or sessions
- Batch processing configuration UI
- API sync with earshot backend
- Windows support

## Documents

- [user-journey.md](user-journey.md) — end-to-end user journey and screen flow
- [../ux-standards.md](../ux-standards.md) — UX principles and interaction patterns
- [../engineering-principles.md](../engineering-principles.md) — engineering principles
- [devices.md](devices.md) — registering and connecting earshot devices
- [recordings.md](recordings.md) — discovery, states, and disposition
- [processing.md](processing.md) — transcription and diarization pipeline
- [output.md](output.md) — transcript format and storage
- [non-functional.md](non-functional.md) — performance and constraints
- [open-questions.md](open-questions.md) — unresolved decisions
