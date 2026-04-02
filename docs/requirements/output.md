# Output Requirements

## Transcript Format

- OUT-1: Transcripts are written as Markdown files named `transcript.md`.
- OUT-2: Transcript structure (v1 — no speaker labels):

```markdown
# Recording — 2026-03-31 09:14:22

**Device:** ritchie-pi
**Duration:** 3m 42s
**Processed:** 2026-03-31 11:05:14

---

[00:03] The meeting is called to order.

[00:11] Thanks everyone for joining.

[00:18] Let's start with the first agenda item.
```

- OUT-3: Timestamps use `[HH:MM:SS]` for recordings over one hour,
  `[MM:SS]` for recordings under one hour.
- OUT-4: Transcripts contain timestamped segments with no speaker labels. Speaker diarization is in the backlog (see docs/backlog.md).

## Storage

- OUT-5: Each recording folder contains all related files together:

```
~/.local/share/earshot-tui/recordings/
  2026-03-31_09-14-22/
    status.json       ← recording state
    recording.opus    ← present if downloaded
    transcript.md     ← present if transcribed
```

- OUT-6: The recordings directory is not configurable. It always resolves to
  `{AppDirs.Data}/recordings/` (`~/.local/share/earshot-tui/recordings/` on
  Linux, `~/Library/Application Support/earshot-tui/recordings/` on macOS —
  see ADR-0005).
- OUT-7: Deleting local audio is deferred to v2.
