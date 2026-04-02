# ADR-0006: JSON Config File for Settings and Device Sources

**Status:** Accepted

## Context

earshot-tui originally stored all configuration (HuggingFace token, setup
completion flag) in the SQLite `config` key-value table, and stored registered
devices in the SQLite `devices` table. This coupled user-editable settings to
the application's internal database, making it awkward to inspect or hand-edit
configuration without a SQLite client.

Device source paths in particular are human-meaningful — they tie a named device
(e.g. `Pi4-Earshot`) to a mount point (e.g. `/run/media/ritchie/EARSHOT`). This
is configuration, not operational state, and belongs in a file the user can read
and modify directly.

## Decision

Move all user-editable settings and device source paths to a JSON file at
`AppDirs.Config()/config.json` (i.e. `~/.config/earshot-tui/config.json`).

### Format

```json
{
  "device_sources": {
    "Pi4-Earshot": "/run/media/ritchie/EARSHOT"
  }
}
```

### Field semantics

| Field | Type | Description |
|-------|------|-------------|
| `device_sources` | object | Map of **device name → host mount path**; each entry is a registered earshot device. At least one entry required. |

### Minimum valid config

At least one `device_sources` entry must be present for the app to function.
If `config.json` does not exist or `device_sources` is absent or empty, the
app launches the setup wizard before proceeding. See setup-wizard.md.

Transcripts are written to the same folder as their recording — there is no
configurable destination (see output.md).

There is no `setup_complete` flag — config validity is determined by inspecting
the fields directly. There is no `huggingface_token` field — speaker diarization
requiring a HuggingFace token is deferred to the backlog (see docs/backlog.md).

`device_sources` is keyed by the device's **human-readable name** (the hostname
read from the device's `preferred_hostname` file, or derived at registration
time). This makes the file directly editable without knowledge of internal UUIDs.

If a device name appears in `device_sources` but its path is not a valid earshot
mount at startup, the entry is silently skipped (the device is offline). If a
device name has no entry at all, the app does not fall back or guess — the user
must add the path manually or run the scan flow.

## Consequences

- SQLite is not used. Recording state is stored in per-recording `status.json`
  files (see recordings.md).
- `config.json` is the only persistent app state outside of recording folders.
- Device sources are version-controlled friendly — users can commit or back up
  `config.json` without including recording history.
- No fallback when source path is missing: the app exits with an error and
  non-zero exit code if a configured device has no `device_sources` entry and no
  mount is found during the scan flow. This is intentional — silent fallbacks
  masked misconfiguration.
- The config file is written atomically (write to temp, rename) to prevent
  partial writes corrupting state.
