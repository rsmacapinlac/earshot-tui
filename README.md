# earshot-tui

A terminal interface for processing [earshot](https://github.com/rsmacapinlac/earshot) recordings locally. Connect your earshot device, download recordings, and get speaker-labeled transcripts — all without sending audio to an external service.

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
- A [HuggingFace](https://huggingface.co) account and access token (required once for the diarization model)

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

## First run

```bash
earshot-tui
```

On first launch, the app will:

1. Detect Python on your system and set up a local processing environment
2. Download the transcription and diarization models (~500MB)
3. Ask for your HuggingFace token to access the diarization model
4. Scan for a connected earshot device

This takes a few minutes once. Subsequent launches are fast.

## Usage

Plug in your earshot device, then launch the app. It will detect the device and show any new recordings. For each recording you can:

- `D` — download and queue for processing
- `S` — skip (leave on device, don't download)
- `X` — delete from device
- `A` — download all remaining

Processing starts automatically after you've made your choices. When complete, open transcripts in your default editor (`$EDITOR`).

To identify speakers, come back to any processed recording, play the audio with `P`, then rename with `R`.

### Key bindings

| Key | Action |
|-----|--------|
| `↑` / `↓` or `j` / `k` | Navigate |
| `Enter` | Select |
| `ESC` | Back / cancel |
| `q` | Quit |
| `C` | Cancel processing |

## Transcripts

Transcripts are saved as Markdown files:

```markdown
# Recording — 2026-03-31 09:14:22

**Device:** earshot-pi
**Duration:** 3m 42s
**Processed:** 2026-03-31 11:05:14

---

[00:03] **Alice:** The meeting is called to order.

[00:11] **Bob:** Thanks everyone for joining.
```

Default location: see `AppDirs` in your platform's standard application data directory.

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
