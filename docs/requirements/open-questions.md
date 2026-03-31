# Open Questions

## Device & Connection

- OQ-1: What is the exact directory structure on the earshot device?
  (confirm `earshot.db` path and `recordings/` layout against earshot source)
- OQ-2: Which fields in the earshot SQLite database must be updated when a
  recording is deleted via the TUI? (recordings table? uploads table? events?)
- OQ-3: How is the device hostname read from the mounted filesystem?
  (e.g., `/etc/hostname` on the mounted volume)

## Processing

- OQ-5: Should the user be able to select model size, or is one model
  baked in for v1? (ADR-0004 defaults to `base` but leaves this open)
- OQ-7: Should inaudible or low-confidence segments appear in the transcript
  with a marker (e.g., `[inaudible]`) or be omitted?

## Output

- OQ-8: Should transcripts include a per-segment confidence score?
- OQ-9: Is the default output directory (`~/earshot-tui/`) the right choice,
  or should it be configurable on first run?

## First Run

- OQ-13: The first-run experience is an undefined feature that must be
  designed before implementation begins. It must cover: Python detection and
  version validation, venv creation, dependency installation, model download,
  HuggingFace token collection, and model license acceptance. Each step needs
  a screen design, progress indication, and a specific recovery path on
  failure. See engineering-principles.md §7.

## Future Considerations

- OQ-10: Drop folder / watch folder support (v2?): should `.opus` files
  manually copied to a folder be processable without a connected device?
- OQ-11: Speaker profiles across recordings (v2?): should naming "Speaker 1"
  as "Alice" on one recording offer to pre-fill on future recordings from the
  same device?
