# ADR-0004: faster-whisper and pyannote.audio as Processing Stack

**Status:** Accepted

## Context

The Python processor (ADR-0003) requires a transcription engine and a
diarization library. The choices here directly affect accuracy, speed,
model download size, and first-run experience.

### Transcription candidates

| Library | Engine | Speed vs Whisper | Notes |
|---|---|---|---|
| openai-whisper | PyTorch | 1× (baseline) | Large install, slow |
| faster-whisper | CTranslate2 | ~4× faster | Smaller memory footprint, binary wheels |
| whisper.cpp (Python bindings) | C++ | Fast | Less Pythonic, fewer integrations |

### Diarization candidates

| Library | Notes |
|---|---|
| pyannote.audio | Industry standard, highest accuracy, requires HuggingFace token |
| simple-diarizer | No token required, lower accuracy |
| WhisperX (wraps both) | Convenient but adds abstraction layer over libraries we'd use directly |

## Decision

Use **faster-whisper** for transcription and **pyannote.audio** for diarization.

For the HuggingFace token requirement: the user accepts the pyannote model
license on HuggingFace once and provides the token to earshot-tui on first
run. The token is stored in the local app config.

## Consequences

- **faster-whisper** ships as binary wheels for macOS, Linux, and Windows —
  no compilation required. Memory usage is significantly lower than
  openai-whisper, which matters for long recordings.
- Model size is configurable. v1 ships with `base` as the default
  (~150MB, good balance of speed and accuracy). The model choice is an
  open question for later resolution (OQ-5).
- **pyannote.audio** provides the best diarization accuracy available in
  open source. The HuggingFace token is a one-time setup step, handled
  during first-run setup alongside venv creation.
- The token is stored in the local SQLite `config` table (see ADR-0006)
  and passed to the processor at runtime via environment variable — it is
  never written to the embedded script or logged.
- For v1, diarization assumes up to 2 speakers (PROC-6). pyannote supports
  `max_speakers=2` to enforce this constraint.
- Both libraries are well-maintained, widely used, and have stable APIs.
  If either is replaced in a future version, the processor interface contract
  (ADR-0003) isolates the change to `processor.py` only.
