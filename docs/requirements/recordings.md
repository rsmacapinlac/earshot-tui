# Recordings Requirements

## Overview

After connecting a device, the app scans for new recordings and presents
them to the user for disposition. The user acts on each recording individually,
in their preferred priority order.

## Recording States (local)

| State       | Description                                                  |
|-------------|--------------------------------------------------------------|
| new         | Present on device, not yet seen by the TUI                   |
| downloaded  | Copied to local storage, queued for processing               |
| processing  | Transcription/diarization in progress                        |
| processed   | Transcript generated successfully                            |
| failed      | Processing attempted but failed                              |
| interrupted | Processing was cancelled or app closed mid-process           |
| skipped     | User chose not to download this session                      |

## Discovery

- REC-1: On device connection, the app scans the device for all `.opus`
  recordings and cross-references against the local database.
- REC-2: Recordings not previously seen are surfaced as `new`.
- REC-3: New recordings are presented as a list showing: timestamp,
  duration. Sorted by most recent first by default.
- REC-4: The user may reorder the list to set processing priority before
  acting on recordings. Reordering uses `alt+↑` / `alt+↓` on the focused
  row. The list component does not support drag-to-reorder.
- REC-5: Previously seen recordings are not re-prompted but are accessible
  in the library.

## Disposition

For each new recording, the user chooses one of three actions:

- REC-6: **[D] Download** — copy the `.opus` file to local storage. After
  all dispositions are set, processing begins automatically in chosen order.
- REC-7: **[S] Skip** — record as skipped; file remains on the device.
  Skipped recordings may be re-encountered if still on the device next session.
- REC-8: **[X] Delete from device** — requires confirmation (see
  ../ux-standards.md §4). Removes `.opus` file from device and updates the
  earshot SQLite database on the device. This is a firm requirement — the
  device DB must be updated to prevent earshot from behaving incorrectly.

Additional:
- REC-9: **[A] Download all** — queues all remaining undecided recordings
  for download in current list order.

## Post-Download: Device Cleanup

- REC-10: After a recording is downloaded and processed, the user may
  delete the local audio file to save space. This is a per-recording,
  explicit action — not automatic.
- REC-11: Deleting local audio does not affect the transcript or recording
  state in the local database.
- REC-12: The device copy of a downloaded recording is left untouched
  unless the user explicitly deletes it (REC-8).
