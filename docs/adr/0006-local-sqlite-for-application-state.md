# ADR-0006: Local SQLite for Application State

**Status:** Accepted

## Context

earshot-tui needs to persist state across sessions: registered devices,
recording history, processing status, speaker names, and HuggingFace token.
This state must survive app restarts and support recovery from interrupted
processing (NFR-6, NFR-9).

The earshot device itself also uses SQLite (`earshot.db`). earshot-tui
maintains a separate local database — it does not extend the device database.

Options considered:

- **SQLite**: Zero-configuration, single file, no server. Well-supported in
  Go via `modernc.org/sqlite` (pure Go, no CGo required).
- **Flat files (JSON/TOML)**: Simple but no atomic updates, harder to query
  across recordings, no transaction support for interrupted-state recovery.
- **Embedded key-value (bbolt)**: Good for simple key-value but awkward for
  relational queries across devices, recordings, and segments.

## Decision

Use **SQLite** via `modernc.org/sqlite` (pure Go driver, no CGo) for all
local application state.

## Schema

```sql
CREATE TABLE devices (
    id          TEXT PRIMARY KEY,   -- UUID
    name        TEXT NOT NULL,
    mount_path  TEXT NOT NULL,
    hostname    TEXT,
    created_at  DATETIME NOT NULL
);

CREATE TABLE recordings (
    id              TEXT PRIMARY KEY,  -- UUID
    device_id       TEXT NOT NULL REFERENCES devices(id),
    device_path     TEXT NOT NULL,     -- path on device at time of discovery
    local_path      TEXT,              -- path to local .opus file, null if not downloaded
    transcript_path TEXT,              -- path to .md file, null if not processed
    timestamp       DATETIME NOT NULL, -- recording start time from filename
    duration_secs   REAL,
    status          TEXT NOT NULL,     -- new|downloaded|processing|processed|failed|interrupted|skipped
    speaker_count   INTEGER,           -- null until processed
    speaker_1_name  TEXT,              -- null until renamed
    speaker_2_name  TEXT,              -- null until renamed
    error_message   TEXT,              -- populated on failed status
    discovered_at   DATETIME NOT NULL,
    processed_at    DATETIME
);

CREATE TABLE config (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
-- Stores: huggingface_token, output_dir, audio_dir, whisper_model
```

## Consequences

- `modernc.org/sqlite` is a pure Go implementation — no CGo, no system
  SQLite dependency, works in cross-compiled binaries for all target platforms.
- The database lives at `AppDirs.Data()/earshot-tui.db`.
- On launch, the app runs migrations automatically. Schema versions are tracked
  in the `config` table (`schema_version` key).
- Interrupted processing recovery (NFR-9): on launch, any recording in
  `processing` status is transitioned to `interrupted`. The user is notified
  and offered a retry.
- earshot-tui **never writes to the earshot device's `earshot.db`** except
  when the user explicitly deletes a recording (REC-8, REC-11) — and only
  the specific fields required to keep the device consistent.
- Speaker names (speaker_1_name, speaker_2_name) are stored per-recording
  and do not propagate across recordings (PROC-12).
- The `config` table stores the HuggingFace token rather than a separate
  config file, keeping all persistent state in one place.
