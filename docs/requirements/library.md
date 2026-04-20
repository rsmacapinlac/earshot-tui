# Library Screen Requirements

## Overview

The library is the main screen of the app. It shows all recordings on the
local machine and allows the user to transcribe recordings, enrich them with
calendar metadata, open transcripts, and navigate back to import. It persists
across sessions — every folder ever imported appears here.

## Entry Points

- `[l] library` from the import screen

## Exit Points

- `[b] back` → import screen (idle only)
- `[q] quit` → exit app (idle only)
- `[enter]` on transcribed row → `$EDITOR` (TUI suspends, resumes on exit)

## Layout

Idle:

```
  Library

    Date        Time    Duration    Status        Progress      Title
  ──────────────────────────────────────────────────────────────────
  > 2026-04-20  10:59   27m 40s     Transcribed                 HLBC Weekly
    2026-04-20  10:03   54m 38s     Encoded
    2026-04-17  12:04   29m 27s     Processed                   HLBC New Shows

  ──────────────────────────────────────────────────────
  [t] transcribe   [e] enrich   [enter] open
  ──────────────────────────────────────────────────────
  [b] back   [q] quit
```

During processing:

```
  Library

    Date        Time    Duration    Status        Progress      Title
  ──────────────────────────────────────────────────────────────────
  > 2026-04-20  10:59   27m 40s     Transcribed                 HLBC Weekly
    2026-04-20  10:03   54m 38s     Transcribing  [████████░░]
    2026-04-17  12:04   29m 27s     Waiting

  ──────────────────────────────────────────────────────
  [t] queue   [e] enrich   [c] cancel
```

## Table Columns

| Column   | Width     | Description |
|----------|-----------|-------------|
| (cursor) | 2         | `>` on the focused row, blank otherwise |
| Date     | 12        | `YYYY-MM-DD` derived from `recorded_at` |
| Time     | 8         | `HH:MM` derived from `recorded_at` |
| Duration | 12        | Human-readable, e.g. `27m 40s`. `—` if unknown |
| Status   | 14        | Text label (see states table below) |
| Progress | 14        | `[████████░░]` during active transcription; blank otherwise |
| Title    | remaining | Enriched meeting title; blank if not yet enriched |

## Status Labels

| Recording state | Label shown  | Phase during transcription |
|-----------------|--------------|---------------------------|
| `new`           | New          | —                         |
| `downloaded`    | Downloaded   | —                         |
| `encoded`       | Encoded      | —                         |
| `processing`    | Processing…  | —                         |
| `transcribed`   | Transcribed  | —                         |
| `processed`     | Processed    | —                         |
| `failed`        | Failed       | —                         |
| `interrupted`   | Interrupted  | —                         |
| queued in memory| Waiting…     | —                         |
| active (loading)| Loading…     | loading                   |
| active (xscribe)| Transcribing | transcribing (+ progress) |
| active (finish) | Completing…  | complete                  |

## Folder List

- LIB-1: All locally imported recordings are listed, sorted most recent first.
- LIB-2: Each row shows date, time, duration, status, progress (when active),
  and enriched title (when available).
- LIB-3: The `>` cursor marker identifies the focused row. Navigation uses
  `[up]`/`[k]` and `[down]`/`[j]`.
- LIB-4: If the library is empty, the screen reads: "No recordings yet.
  Press [b] to go back and import from a device." Only `[b] back` and
  `[q] quit` are shown.

## Transcription

- LIB-5: `[t] transcribe` begins transcription of the focused recording.
  There are no preconditions on recording state — any recording can be
  (re-)transcribed.
- LIB-6: While a transcription is in progress, pressing `[t]` on any other
  row adds it to the queue. Queued rows show "Waiting…" in the Status column.
- LIB-7: Jobs run one at a time in the order they were queued.
- LIB-8: The active row shows the current phase in Status (`Loading…`,
  `Transcribing`, `Completing…`) and a visual progress bar in the Progress
  column during the transcribing phase.
- LIB-9: When a recording completes, its status updates to `Transcribed` and
  the next queued recording starts automatically.
- LIB-10: When a recording fails, its status updates to `Failed`. Remaining
  queued recordings continue unaffected.

## Calendar Enrichment

- LIB-11: `[e] enrich` looks up the focused recording's `recorded_at`
  timestamp against ICS calendar files found under `config.calendar_dir`.
- LIB-12: On a match, the recording's `status.json` is updated with the
  calendar event's title, attendees, and description. The Title column
  updates to show the meeting title.
- LIB-13: If no matching event is found, or `calendar_dir` is not configured,
  an inline error is shown.
- LIB-14: Enrichment is available in both idle and processing states.

## Opening a Transcript

- LIB-15: `[enter]` opens the transcript for the focused recording in
  `$EDITOR` using `tea.ExecProcess()` — the TUI suspends and resumes when
  the editor exits. The library reloads after the editor closes.
- LIB-16: `[enter]` is only active when the focused recording is in
  `transcribed` or `processed` state.

## Cancellation

- LIB-17: `[c] cancel` is only shown during active transcription.
- LIB-18: On cancel, the in-progress recording is marked `interrupted`.
  All queued recordings have their queued flag cleared (their persisted
  state is unchanged).

## Navigation

- LIB-19: `[b] back` navigates to the import screen. Only available when
  no transcription is in progress.
- LIB-20: `[q] quit` exits the app. Only available when no transcription
  is in progress.
