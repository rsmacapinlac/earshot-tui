# ADR-0004: faster-whisper as Transcription Engine

**Status:** Accepted

## Context

The Python processor (ADR-0003) requires a transcription engine. The choice
here directly affects accuracy, speed, model download size, and first-run
experience.

### Transcription candidates

| Library | Engine | Speed vs Whisper | Notes |
|---|---|---|---|
| openai-whisper | PyTorch | 1× (baseline) | Large install, slow |
| faster-whisper | CTranslate2 | ~4× faster | Smaller memory footprint, binary wheels |
| whisper.cpp (Python bindings) | C++ | Fast | Less Pythonic, fewer integrations |

## Decision

Use **faster-whisper** for transcription. Speaker diarization is in the
backlog (see docs/backlog.md) and not part of v1.

No HuggingFace token is required. faster-whisper downloads its models from
public (ungated) HuggingFace repositories automatically.

## Consequences

- **faster-whisper** ships as binary wheels for macOS, Linux, and Windows —
  no compilation required. Memory usage is significantly lower than
  openai-whisper, which matters for long recordings.
- Model size is fixed at `base` (~150MB, good balance of speed and accuracy).
  Model selection is a backlog consideration.
- v1 produces transcripts with segment-level timestamps and no speaker labels.
- No HuggingFace token, no license acceptance step, no user input required
  during preflight.
- When speaker diarization is added from the backlog, the processor interface
  contract (ADR-0003) isolates the change to `processor.py` only.
