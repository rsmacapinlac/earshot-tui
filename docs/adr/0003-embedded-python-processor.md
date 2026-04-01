# ADR-0003: Embedded Python Processor

**Status:** Accepted

## Context

earshot-tui requires local transcription. The best available library
(faster-whisper) is Python-native with no mature equivalent in Go (see ADR-0004).

Rather than writing the TUI in Python (which complicates distribution and
concurrency) or using Docker (which adds a heavyweight dependency and poor
first-run UX), a clean separation is established:

- **Go** owns everything the user interacts with.
- **Python** owns everything about understanding audio.

The processing component is abstracted entirely from the user — they install
one binary and it works.

Alternatives considered:

- **Docker-hosted processor**: Requires Docker installed and running.
  No similar transcription tool uses Docker as a hard dependency. Poor
  first-run UX.
- **Go-native whisper.cpp bindings**: Fewer Python integrations and lower
  accuracy than faster-whisper. Diarization (backlog) would still require
  Python regardless.
- **Separate Python package (pip install)**: Exposes the Python layer to the
  user. Adds installation steps and version management burden.

## Decision

Bundle `processor.py` and `requirements.txt` inside the Go binary using
`//go:embed`. At runtime, Go manages the Python environment invisibly.

## Runtime Behaviour

```
First launch:
  1. Extract processor.py + requirements.txt to UserConfigDir/processor/
  2. Locate Python 3.10+ on host (via PythonResolver — see ADR-0005)
  3. Create venv at UserConfigDir/venv/
  4. pip install -r requirements.txt
  5. Download faster-whisper base model to UserCacheDir/huggingface/

Subsequent launches:
  1. Hash embedded requirements.txt against installed state
  2. Match → proceed immediately
  3. Mismatch (binary updated) → re-run pip install, then proceed

Processing a recording:
  Go spawns: venv/bin/python processor.py /path/to/audio.opus
  Python writes progress to stderr  → Go reads → updates progress bar
  Python writes result to stdout    → Go reads → parses JSON → writes Markdown
```

## Interface Contract

**Input:** absolute path to `.opus` file passed as CLI argument.

**stdout:** single JSON object on completion:
```json
{
  "version": "1",
  "duration": 222.4,
  "segments": [
    {
      "start": 0.5,
      "end": 3.1,
      "text": "The meeting is called to order."
    }
  ]
}
```

**stderr:** progress lines read by Go for the progress bar:
```
PROGRESS:transcribing:0.45
PROGRESS:complete
ERROR:Could not load model: ...
```

**Exit codes:** `0` success, `1` processing error (details on stderr).

## Consequences

- **Python is a host dependency.** The app cannot function without Python 3.10+
  installed. If not found, a specific error is shown with platform-appropriate
  install instructions (../ux-standards.md §6).
- The binary is self-contained for all other purposes — no separate files to
  distribute or manage.
- When `processor.py` or `requirements.txt` change, the entire Go binary is
  recompiled and redistributed. The venv is automatically refreshed on next
  launch via requirements hash check.
- The processor is independently testable: run `python processor.py audio.opus`
  directly against any `.opus` file.
- The processor can be replaced or upgraded (different model, different
  library) with no changes to the Go codebase, provided the interface
  contract above is maintained.
- On Windows, `venv/bin/python` becomes `venv\Scripts\python.exe` — resolved
  transparently by the PythonResolver (ADR-0005).
