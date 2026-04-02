#!/usr/bin/env python3
"""
earshot-tui processor — transcribes a .opus audio file using faster-whisper.

Interface contract (see docs/adr/0003-embedded-python-processor.md):
  Input:  path to .opus file passed as sys.argv[1]
  stdout: single JSON object on completion
  stderr: PROGRESS:<phase>:<fraction> lines, ERROR:<message> on failure
  Exit:   0 = success, 1 = error
"""

import json
import os
import sys


def main() -> None:
    if len(sys.argv) < 2:
        print("ERROR:Usage: processor.py <audio.opus>", file=sys.stderr, flush=True)
        sys.exit(1)

    audio_path = sys.argv[1]

    if not os.path.exists(audio_path):
        print(f"ERROR:File not found: {audio_path}", file=sys.stderr, flush=True)
        sys.exit(1)

    # Stub mode: return fake output without loading any model.
    if os.environ.get("EARSHOT_PROCESSOR_STUB") == "1":
        _run_stub()
        return

    cache_dir = os.environ.get("EARSHOT_CACHE_DIR")

    try:
        from faster_whisper import WhisperModel
    except ImportError as exc:
        print(f"ERROR:Failed to import faster_whisper: {exc}", file=sys.stderr, flush=True)
        sys.exit(1)

    try:
        print("PROGRESS:loading:0.0", file=sys.stderr, flush=True)
        model = WhisperModel("base", device="cpu", download_root=cache_dir)

        print("PROGRESS:transcribing:0.0", file=sys.stderr, flush=True)
        segments_iter, info = model.transcribe(audio_path, beam_size=5)

        segments = []
        total = info.duration

        for seg in segments_iter:
            text = seg.text.strip()
            segments.append({"start": seg.start, "end": seg.end, "text": text})
            if total > 0:
                fraction = min(seg.end / total, 1.0)
                print(f"PROGRESS:transcribing:{fraction:.3f}", file=sys.stderr, flush=True)

        print("PROGRESS:complete", file=sys.stderr, flush=True)
        result = {"version": "1", "duration": total, "segments": segments}
        print(json.dumps(result))

    except Exception as exc:
        print(f"ERROR:{exc}", file=sys.stderr, flush=True)
        sys.exit(1)


def _run_stub() -> None:
    """Return deterministic fake output without loading any ML model."""
    print("PROGRESS:transcribing:0.5", file=sys.stderr, flush=True)
    print("PROGRESS:complete", file=sys.stderr, flush=True)
    result = {
        "version": "1",
        "duration": 10.0,
        "segments": [
            {"start": 0.5, "end": 3.0, "text": "This is a stub transcription."},
            {"start": 3.5, "end": 6.0, "text": "The processor is running in stub mode."},
        ],
    }
    print(json.dumps(result))


if __name__ == "__main__":
    main()
