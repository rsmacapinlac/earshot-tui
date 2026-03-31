# Engineering Principles

Guiding principles for the development of earshot-tui. These exist because
the architecture has specific risks — a long external dependency chain, a
Go/Python process boundary, irreplaceable user data, and a complex first-run
experience. Each principle directly guards against a known failure mode.

When a decision is unclear, check here first.

---

## 1. Fail at the Gate, Not in the Field

All external dependencies — Python version, ffmpeg, HuggingFace token, model
availability — must be verified at startup before the user enters any workflow.
A user who downloads three recordings and then hits a missing `ffmpeg` error
during processing has lost trust in the app. Dependency failures must surface
before the user invests time.

**Guards against:** Silent failures surfacing at the worst possible moment.

**In practice:**
- Startup performs a dependency preflight check in sequence
- Each failure produces a specific, actionable error (see Principle 10)
- The app does not proceed past the preflight until all checks pass

---

## 2. The Processor Contract Is an API

The JSON interface between the Go binary and the Python processor (ADR-0003)
is a versioned, formal API — not an internal detail. Both sides validate
against it explicitly. The processor embeds a `version` field in its output.
Breaking changes to the contract require a version bump and a corresponding
update to the Go consumer.

**Guards against:** Go/Python boundary bugs that are hard to diagnose and
reproduce across environments.

**In practice:**
- `processor.py` outputs `{"version": "1", "duration": ..., "segments": [...]}`
- Go validates the version field before parsing the rest
- Changes to the output schema are treated as API changes, not refactors

---

## 3. User Data Is Never at Risk

No operation leaves the user's local database, transcripts, or audio files in
an inconsistent state. Every write operation that can fail midway must be
atomic. The earshot device database is never written to without explicit,
confirmed user intent.

**Guards against:** Data loss or corruption of recordings that may be
irreplaceable.

**In practice:**
- File writes use a temp file + rename pattern (atomic on POSIX)
- Database writes that span multiple rows use transactions
- A partial download is never marked `downloaded` — the status is set only
  after the file is fully written and verified
- Writes to the earshot device `earshot.db` happen only on explicit delete
  actions (REC-7, REC-8), never speculatively

---

## 4. Platform Differences Live in One Place

No platform-specific logic exists outside `internal/platform/`. This is a
hard rule, not a guideline. The four interfaces defined in ADR-0005
(`PythonResolver`, `MountScanner`, `AudioPlayer`, `AppDirs`) are the only
permitted entry points for platform behaviour. A platform check anywhere else
in the codebase is a bug.

**Guards against:** Windows support becoming a codebase-wide refactor instead
of a contained implementation task.

**In practice:**
- Code review rejects any `runtime.GOOS` check outside `internal/platform/`
- All platform implementations are covered by interface compliance tests
- Platform-specific file paths, commands, and APIs are constants defined in
  platform files, never inline strings

---

## 5. The Processor Is Always Independently Runnable

`processor.py` must function as a standalone script at all times:

```bash
python processor.py /path/to/audio.opus
```

No dependency on the Go binary. No assumption about how it was invoked. Valid
JSON to stdout, progress lines to stderr, correct exit codes. This is both a
testing requirement and a user escape hatch for debugging.

**Guards against:** Untestable processing code and environment-specific
failures with no diagnostic path.

**In practice:**
- `processor.py` is tested in CI independently of the Go binary
- The processor reads its config (HuggingFace token, model size) from
  environment variables so it can be invoked with any environment
- The interface contract (ADR-0003) is validated by a standalone test suite
  against real `.opus` fixture files

---

## 6. stderr from Subprocesses Is Untrusted

Go never parses unstructured stderr from the Python processor. Only lines
explicitly matching the `PROGRESS:` protocol are acted upon. Everything else
is written to the debug log and ignored. PyTorch, pyannote, and HuggingFace
libraries produce unpredictable, version-dependent stderr output that must
not influence application behaviour.

**Guards against:** Library version-dependent fragility in the progress
reporting path.

**In practice:**
- The Go stderr reader filters for `^PROGRESS:` and `^ERROR:` prefixes only
- All other stderr lines are written to the debug log (see Principle 11)
- `processor.py` suppresses library warnings at the Python level where
  possible (`warnings.filterwarnings`, logging configuration)
- The protocol is documented in ADR-0003 and treated as part of the API

---

## 7. First Run Is a Defined Feature

The first-run experience — Python detection, venv creation, model download,
HuggingFace token collection, license acceptance — is a user experience, not
an installation side effect. It requires its own screen, real progress
indicators, clear explanations for each step, and specific recovery paths
when something goes wrong.

**Guards against:** Abandonment before the app proves its value.

> **Status: Undefined.** The specific UX flow, screen design, and step
> sequence for first run has not yet been designed. This is a required feature
> for v1 and must be resolved before implementation begins. See OQ-13.

---

## 8. Recovery Is a First-Class Feature

Interrupted state — app closed mid-processing, device disconnected
mid-download, pip install timed out — is not an edge case. Every operation
that can be interrupted has a defined recovery path. On next launch, the app
reconciles state and offers clear options. It never silently resumes, silently
fails, or presents stale state as current.

**Guards against:** Interruptions creating permanent broken state that erodes
trust in the app.

**In practice:**
- On launch, recordings in `processing` state are transitioned to
  `interrupted` and the user is notified (PROC-20)
- Partial downloads are detected (file size vs. expected size) and discarded;
  the recording returns to `new` state
- A broken venv (failed pip install) is detected at startup and triggers
  a clean rebuild, not a silent failure

---

## 9. requirements.txt Is a Tested Artifact

`requirements.txt` is a pinned, tested lockfile — not a wish list. It is
validated on every supported platform before being embedded in a release
binary. Loose version pins (`>=`) are not permitted for direct dependencies.
PyTorch and pyannote have a documented history of subtle cross-version
incompatibilities.

**Guards against:** Cross-platform dependency failures that reproduce only
in user environments.

**In practice:**
- CI runs `pip install -r requirements.txt` and executes the processor test
  suite on macOS arm64, macOS x86-64, Linux x86-64, and Linux arm64
- `requirements.txt` uses exact pins (`==`) for all direct dependencies
- Dependency updates are treated as releases: tested fully before embedding

---

## 10. Errors Identify a Cause and an Action

Every error surfaced to the user names what went wrong and what they can do
next. No raw exception messages. No "something went wrong." This applies to
the Go layer, the processor's `ERROR:` stderr lines, and first-run setup
output. The error format in ../ux-standards.md §6 is mandatory.

**Guards against:** User frustration and unactionable support requests in an
app with many environment-specific failure modes.

**In practice:**
- Each known error condition has a defined message written at the point the
  error is detected, not where it is caught
- Python `ERROR:` lines include a code that maps to a user-facing message in
  the Go layer — raw Python tracebacks are never shown to the user
- Unknown errors fall back to: "An unexpected error occurred. Run with
  `--debug` for details." (see Principle 11)

---

## 11. Diagnostics Are Built In

The app provides a `--debug` flag that writes a structured log of the full
subprocess lifecycle: Python path resolved, venv path, pip output, all stderr
from the processor, and JSON received. This log is what a user attaches to a
bug report. Without it, debugging environment-specific failures is guesswork.

**Guards against:** Support requests that cannot be diagnosed remotely.

**In practice:**
- `earshot-tui --debug` writes to `stderr` and to a log file at
  `AppDirs.Data()/debug.log`
- The debug log includes: platform info, Python version found, venv path,
  full pip output on setup, all subprocess stderr, and the raw JSON received
  from the processor
- Debug mode is never on by default — it may log paths to audio files

---

## 12. Subprocess Lifecycle Is Explicitly Owned

Every child process spawned by the app — the Python processor, the audio
player — has an explicit owner responsible for its full lifecycle: start,
monitor, stop, and clean up on exit. No child process is left running if
the TUI exits, the user cancels, or an error occurs.

**Guards against:** Orphaned processes consuming resources after the app closes.

**In practice:**
- Go uses process groups for the Python processor so that all child threads
  (PyTorch workers) are terminated together on cancel (PROC-15)
- `AudioPlayer` implementations manage their player process and terminate it
  if the TUI exits while audio is playing
- A deferred cleanup function is registered at startup to kill any running
  child processes on app exit
- On Windows (future): `TerminateProcess` is used in place of `SIGTERM`
