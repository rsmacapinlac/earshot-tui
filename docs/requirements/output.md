# Output Requirements

## Transcript Format

- OUT-1: Transcripts are written as Markdown files.
- OUT-2: Each transcript is named after the recording timestamp:
  `YYYY-MM-DD_HH-MM-SS.md`
- OUT-3: Transcript structure:

```markdown
# Recording — 2026-03-31 09:14:22

**Device:** ritchie-pi
**Duration:** 3m 42s
**Processed:** 2026-03-31 11:05:14
**Speakers:** 2 detected
> Note: Speaker detection is limited to 2 speakers in v1. Rename speakers
> from the recording detail screen after listening to the audio.

---

[00:03] **Speaker 1:** The meeting is called to order.

[00:11] **Speaker 2:** Thanks everyone for joining.

[00:18] **Speaker 1:** Let's start with the first agenda item.
```

- OUT-4: Timestamps use `[HH:MM:SS]` for recordings over one hour,
  `[MM:SS]` for recordings under one hour.
- OUT-5: Speaker names reflect any renaming applied. Default is "Speaker 1"
  and "Speaker 2".
- OUT-6: The speaker limit note (OUT-3) is only included when 2 speakers are
  detected. Single-speaker recordings have no speaker labels or notes.
- OUT-7: When speakers are renamed, the `.md` file is rewritten in place.
  The note in the header is removed after renaming.

## Storage

- OUT-8: Transcripts are stored in a configurable output directory.
  Default: `{AppDirs.Data}/transcripts/` (resolves to
  `~/Library/Application Support/earshot-tui/transcripts/` on macOS,
  `~/.local/share/earshot-tui/transcripts/` on Linux — see ADR-0005).
- OUT-9: Audio files are stored in a configurable audio directory.
  Default: `{AppDirs.Data}/audio/` (resolves to
  `~/Library/Application Support/earshot-tui/audio/` on macOS,
  `~/.local/share/earshot-tui/audio/` on Linux — see ADR-0005).
- OUT-10: Audio is kept alongside transcripts by default. The user may
  delete local audio per recording via the recording detail screen.
- OUT-11: Deleting local audio does not affect the transcript or local
  database state.
