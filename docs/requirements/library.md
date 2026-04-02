# Library Screen Requirements

## Overview

The library is the main screen of the app. It shows all recordings on the
local machine and allows the user to process downloaded folders, open
transcripts, and navigate to import. It persists across sessions — every
folder ever imported appears here.

## Entry Points

- `[l] library` from the import screen

## Exit Points

- `[b] back` → import screen
- `[q] quit` → exit app
- `[enter]` on transcribed folder → `$EDITOR` (TUI suspends, resumes on exit)

## Layout

Idle:

```
  [✓] 2026-03-31 09:14  Transcribed    8m 45s
  [ ] 2026-03-30 14:22  Downloaded  14m 22s
  [✗] 2026-03-29 08:55  Failed       3m 12s

  ──────────────────────────────────────────────────────
  [space] select   [t] transcribe
  ──────────────────────────────────────────────────────
  [b] back   [q] quit
```

During processing:

```
  [✓] 2026-03-31 09:14  Transcribed    8m 45s
  [✓] 2026-03-30 14:22  Processing…  8m 45s  ████████░░  72%
  [✓] 2026-03-29 08:55  Waiting…
  [✗] 2026-03-28 16:40  Failed       3m 12s

  ──────────────────────────────────────────────────────
  [c] cancel
```

## Folder List

- LIB-1: All locally imported folders are listed, sorted most recent first.
- LIB-2: Each row shows: folder name with timestamp derived from the folder
  name, status, and duration.
- LIB-3: Status is shown with a colour indicator (see ux-standards.md §11):

  | Status      | Marker | Colour  |
  |-------------|--------|---------|
  | Downloaded  | none   | Yellow  |
  | Processing  | `[⠸]`  | Yellow  |
  | Waiting     | none   | Yellow  |
  | Transcribed   | `[✓]`  | Green   |
  | Failed      | `[✗]`  | Red     |
  | Interrupted | `[!]`  | Red     |

- LIB-4: If the library is empty, the screen reads: "No recordings yet.
  Press [b] to go back and import from a device." Only `[b] back` and `[q] quit`
  are shown.

## Selection and Processing

- LIB-5: `[space]` toggles selection on the focused folder. Only folders in
  `downloaded`, `failed`, or `interrupted` state are selectable.
  `transcribed` folders cannot be selected.
- LIB-6: `[t] transcribe` begins transcription of all selected folders in list
  order (most recent first).
- LIB-7: `[t] transcribe` is only shown when at least one folder is selected
  and no transcription is in progress.
- LIB-8: During transcription, the active folder row shows an inline progress
  bar. The checkbox is replaced by `[⠸]`. Queued folders show "Waiting..."
- LIB-9: When a folder completes, its row updates to `[✓] Transcribed`.
- LIB-10: When a folder fails, its row updates to `[✗] Failed`.
- LIB-11: One folder's failure does not block others in the queue.

## Opening a Transcript

- LIB-12: `[enter]` opens the transcript for the focused folder in `$EDITOR`
  using `tea.ExecProcess()` — the TUI suspends and resumes when the editor exits.
- LIB-13: `[enter]` is not shown in the footer but is always active when the
  focused folder is in `transcribed` state. It has no effect on other states.

## Cancellation

- LIB-14: `[c] cancel` is only shown during active transcription.
- LIB-15: On cancel, the in-progress folder is marked `interrupted` and
  remaining queued folders return to their previous state (unselected).

## Navigation

- LIB-16: `[b] back` navigates to the import screen. Hidden during active
  transcription.
- LIB-17: `[q] quit` exits the app. Hidden during active transcription.
- LIB-18: During active transcription, only `[c] cancel` is shown in the
  footer.
