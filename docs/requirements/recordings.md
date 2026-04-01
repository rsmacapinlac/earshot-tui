# Recordings Requirements

## Overview

After connecting a device, the app scans for new recordings and presents
them to the user for import. The user acts on each folder individually
from the import screen.

## Recording States (local)

| State       | Description                                                  | Persisted |
|-------------|--------------------------------------------------------------|-----------|
| new         | Present on device, not yet acted on by the TUI               | No — derived at scan time |
| downloaded  | Copied to local storage, queued for processing               | Yes — `status.json` created |
| processing  | Transcription in progress                                    | Yes — recovery target on next launch |
| completed   | Transcript generated successfully                            | Yes |
| failed      | Processing attempted but failed                              | Yes — offer retry |
| interrupted | App closed mid-processing                                    | Yes — offer retry on next launch |

State is persisted in a `status.json` file within each recording's local folder:

```
~/.local/share/earshot-tui/recordings/
  2026-03-31_09-14-22/
    status.json
    recording.opus    ← present if downloaded
    transcript.md     ← present if completed
```

```json
{
  "status": "completed",
  "device": "Pi4-Earshot",
  "recorded_at": "2026-03-31T09:14:22Z",
  "duration": 222,
  "downloaded_at": "2026-03-31T11:00:00Z",
  "completed_at": "2026-03-31T11:05:14Z"
}
```

There is no `skipped` state. Recordings the user does not import remain on
the device and are re-surfaced on the next session.

## Discovery

- REC-1: On device connection, the app scans the device for all recording
  folders and cross-references against the local recordings directory.
- REC-2: Folders not previously imported are surfaced as `new` on the import
  screen.
- REC-3: New folders are presented showing: timestamp (derived from folder
  name), recording count, total duration. Sorted most recent first.
- REC-4: Previously imported folders are not re-prompted on the import screen
  but are accessible in the library.

## Post-Download: Device Cleanup

- REC-5: The device copy of a downloaded recording is left untouched. There
  is no delete-from-device action in v1.
- REC-6: Deleting local audio files is deferred to v2.
