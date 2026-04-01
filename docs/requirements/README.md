# Earshot TUI — Requirements

A terminal user interface that runs on a desktop/laptop to download, process,
and manage recordings from one or more [earshot](https://github.com/rsmacapinlac/earshot)
devices. Transcription and speaker diarization run entirely locally.

## Scope (v1)

- USB-connected earshot device management (auto-detect, hostname-based naming)
- Local transcription (faster-whisper, no network calls)
- Markdown transcript output with timestamps
- Folder-level import from device
- Library view with inline processing
- No authentication, no API integration

## Out of Scope (v1)

- Speaker diarization and renaming (v2)
- Drop folder / watch folder support (v2)
- Speaker profiles persisted across recordings or sessions (v2)
- API sync with earshot backend
- Windows support

## Documents

- [user-journey.md](user-journey.md) — end-to-end user journey and screen flow
- [../ux-standards.md](../ux-standards.md) — UX principles and interaction patterns
- [../engineering-principles.md](../engineering-principles.md) — engineering principles
- [setup-wizard.md](setup-wizard.md) — setup wizard screen requirements
- [devices.md](devices.md) — registering and connecting earshot devices
- [recordings.md](recordings.md) — recording states and storage
- [import.md](import.md) — import screen requirements
- [library.md](library.md) — library screen requirements
- [processing.md](processing.md) — transcription pipeline
- [output.md](output.md) — transcript format and storage
- [non-functional.md](non-functional.md) — performance and constraints
- [open-questions.md](open-questions.md) — unresolved decisions
