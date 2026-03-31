# Processing Requirements

## Overview

Processing converts a downloaded `.opus` recording into a speaker-labeled
transcript. All processing runs locally on the user's computer. Processing
starts automatically after recordings are downloaded — no manual trigger needed.

## Pipeline

```
.opus file
    → decode audio
    → transcribe (local model) → timestamped segments
    → diarize → speaker count + labeled segments
    → merge transcription + diarization
    → write Markdown transcript
```

## Transcription

- PROC-1: Processing begins automatically once a recording reaches
  `downloaded` state. No user action required.
- PROC-2: Recordings are processed in the order the user chose during
  disposition.
- PROC-3: Transcription runs locally — no audio is sent to any external
  service.
- PROC-4: Transcription produces segment-level timestamps.

## Speaker Diarization

- PROC-5: Diarization runs after transcription and detects the number of
  distinct speakers in the recording.
- PROC-6: For v1, detection supports up to 2 speakers. If more than 2
  speakers are present, segments are collapsed into 2 buckets. This
  limitation is noted in the transcript output (see output.md).
- PROC-7: If only 1 speaker is detected, no speaker labels or renaming
  options are shown. The transcript reads as a single voice.
- PROC-8: Speaker renaming is optional, per-recording, and available only
  when 2 speakers are detected. It is not offered during or immediately
  after processing — the user must listen to the recording first.

## Speaker Renaming

- PROC-9: The user initiates renaming from the recording detail screen after
  listening to the audio.
- PROC-10: Renaming prompts for both speaker names:
  ```
  Speaker 1: [____________]
  Speaker 2: [____________]
  ```
- PROC-11: On confirmation, the app rewrites the `.md` transcript file in
  place, replacing all instances of "Speaker 1" / "Speaker 2" with the
  chosen names.
- PROC-12: Speaker names are stored per-recording in the local database.
  They do not carry over to other recordings.
- PROC-13: The user may re-rename speakers at any time after processing.

## Progress and Cancellation

- PROC-14: A progress bar is shown per recording and for the overall queue.
  The status line shows what step is running (e.g., "Transcribing 2 of 4...").
- PROC-15: `[C] Cancel` is available at all times during processing. On
  cancel, the current recording is marked `interrupted` and remaining
  queued recordings return to `downloaded` state.
- PROC-16: The UI remains navigable during processing (see ../ux-standards.md §5).

## Completion

- PROC-21: When the processing queue is complete, the app displays a summary
  list of all recordings just processed with their status (processed/failed).
- PROC-22: The user selects any processed recording from the summary to open
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
