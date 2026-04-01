# Open Questions

## Device & Connection

- OQ-1: What is the exact directory structure on the earshot device?
  (confirm `earshot.db` path and `recordings/` layout against earshot source)
- OQ-2: Which fields in the earshot SQLite database must be updated when a
  recording is deleted via the TUI? (recordings table? uploads table? events?)
- OQ-3: How is the device hostname read from the mounted filesystem?
  (e.g., `/etc/hostname` on the mounted volume)

## Processing

- OQ-7: Should inaudible or low-confidence segments appear in the transcript
  with a marker (e.g., `[inaudible]`) or be omitted?
- OQ-14: When a folder contains multiple `.opus` files, are they concatenated
  and transcribed as a single audio stream, or transcribed individually and
  merged into one `transcript.md`? This affects the processor contract
  (ADR-0003) and timestamp continuity in the output.

## Output

- OQ-8: Should transcripts include a per-segment confidence score?

## Future Considerations

- OQ-10: Drop folder / watch folder support: should `.opus` files manually
  copied to a folder be processable without a connected device?
