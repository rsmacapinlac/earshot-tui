# Processing Requirements

## Overview

Processing converts a downloaded `.opus` recording into a timestamped
transcript. All processing runs locally on the user's computer.

## Pipeline

```
.opus file
    → decode audio
    → transcribe (local model) → timestamped segments
    → write Markdown transcript
```

## Transcription

- PROC-1: Processing is manually triggered by the user from the library
  screen. The user selects folders and presses `[p] process`.
- PROC-2: Folders are transcribed in the order selected, most recent first.
- PROC-3: Transcription runs locally — no audio is sent to any external
  service.
- PROC-4: Transcription produces segment-level timestamps.

## Progress and Cancellation

- PROC-14: A progress bar is shown per recording and for the overall queue.
  The status line shows what step is running (e.g., "Transcribing 2 of 4...").
- PROC-15: `[C] Cancel` is available at all times during processing. On
  cancel, the current recording is marked `interrupted` and remaining
  queued recordings return to `downloaded` state.
- PROC-16: The UI remains navigable during processing (see ../ux-standards.md §5).

## Completion

- PROC-21: When the processing queue is complete, the app displays a summary
  list of all recordings just completed with their status (completed/failed).
- PROC-22: The user selects any completed recording from the summary to open
  its transcript in `$EDITOR`. Multiple transcripts can be opened in sequence.
- PROC-23: From the summary, the user can also navigate to the library or
  connect another device.

## Error Handling

- PROC-17: If processing fails, the recording is marked `failed` with a
  specific error message displayed in the recording detail.
- PROC-18: The user may retry a `failed` or `interrupted` recording at any time.
- PROC-19: One recording's failure does not block others in the queue.
- PROC-20: If the app is closed during processing, recordings in-progress
  are marked `interrupted` on next launch. The user is notified and offered
  the option to retry.
