# earshot-tui

A terminal interface for processing [earshot](https://github.com/rsmacapinlac/earshot) recordings locally. Connect your earshot device, download recordings, and get transcripts — all without sending audio to an external service.

## What it does

earshot is a Raspberry Pi device that captures conversations. earshot-tui runs on your laptop or desktop and handles what comes next:

- Connects to your earshot device over USB and finds new recordings
- Downloads recordings in the order you choose
- Transcribes and identifies speakers locally on your machine
- Produces Markdown transcripts with timestamps
- Lets you listen to recordings and rename speakers after the fact

## Requirements

- macOS or Linux (Windows planned)
- Python 3.10 or later

> **Note:** earshot-tui ships as a single binary. Python is used internally for transcription and diarization — you do not need to manage any Python packages yourself.

## Installation

Download the latest binary for your platform from the [releases page](#) and place it somewhere on your `$PATH`:

```bash
# macOS (Apple Silicon)
curl -L https://github.com/rsmacapinlac/earshot-tui/releases/latest/download/earshot-tui-darwin-arm64 -o earshot-tui
chmod +x earshot-tui
mv earshot-tui /usr/local/bin/

# Linux (x86-64)
curl -L https://github.com/rsmacapinlac/earshot-tui/releases/latest/download/earshot-tui-linux-amd64 -o earshot-tui
chmod +x earshot-tui
mv earshot-tui /usr/local/bin/
```

## Run from source

For local development and manual testing:

1. Install **Go** (1.22+), **Python 3.10–3.12** (recommended for PyTorch wheels), **ffmpeg** on your `PATH`, and Git.
2. Clone the repo and run:

   ```bash
   go run ./cmd/earshot-tui
   ```

3. Complete the first-run setup (venv, `pip install` from the embedded `requirements.txt`, Hugging Face token). The first install can take several minutes and pulls large ML dependencies.
4. Optional: set `EARSHOT_PROCESSOR_STUB=1` in the environment before launching so the embedded processor returns fake output without running Whisper or pyannote (useful for exercising the TUI and DB flow). Dependencies are still installed if you run the full setup wizard.

## Platform support

| Platform | Status |
|----------|--------|
| macOS (Apple Silicon, Intel) | v1 |
| Linux (x86-64, ARM64) | v1 |
| Windows | Planned |

## Documentation

- [Requirements](docs/requirements/README.md)
- [User journey](docs/requirements/user-journey.md)
- [UX standards](docs/ux-standards.md)
- [Architecture decisions](docs/adr/README.md)
- [Engineering principles](docs/engineering-principles.md)

## Related

- [earshot](https://github.com/rsmacapinlac/earshot) — the Raspberry Pi recording device this tool is built for
