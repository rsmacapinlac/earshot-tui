# Backlog

Features and requirements that are explicitly out of scope for v1. They are
recorded here so the decisions are visible and the work is not lost.

---

## Speaker Diarization

Identify and label individual speakers within a transcript.

**Why deferred:** pyannote.audio requires a HuggingFace token and license
acceptance, adds a significant dependency footprint, and requires solving the
speaker naming UX. These are solvable problems but add scope that would delay
a working v1.

**What this unlocks:**
- Transcript segments annotated with speaker labels (`SPEAKER_00`, `SPEAKER_01`, …)
- Speaker renaming flow: user assigns names to generic labels after processing
- Speaker profiles across recordings: a name assigned to a speaker on one
  recording can pre-fill on future recordings from the same device

**Implementation notes when revisited:**
- pyannote.audio is the target library (industry standard, highest accuracy)
- Requires HuggingFace token — add `huggingface_token` field to `config.json`
  and a token entry step to the setup wizard
- Processor interface contract (ADR-0003) will need `"speakers"` count and
  `"speaker"` field on each segment; this is a version bump
- Speaker renaming is a post-processing step on the Markdown file
  (`transcript/rewrite.go`)
- Speaker profiles could be stored as a `speakers.json` per device in the
  config directory

**Deferred requirements:** PROC-5 through PROC-13, OQ-11

---

## Model Selection

Allow the user to choose a Whisper model size (tiny, base, small, medium, large).

**Why deferred:** Hardcoding `base` keeps first-run download size predictable
(~150MB) and avoids a configuration decision in the setup wizard. Model
selection adds surface area without materially improving v1 utility for most
users.

---

## Local Audio Deletion

Delete local `.opus` files after a transcript has been produced, or on demand.

**Why deferred:** Irreversible action on data that may be the only copy.
Needs a confirmation flow and consideration of what happens to the transcript
if audio is deleted. Not needed for v1 utility.

---

## Processing Spinner Marker

During active transcription the processing row should show `[⠸]` (spinner character) instead of `[✓]`. Currently both the active and queued rows show `[✓]` in yellow, making it hard to distinguish which recording is actively running vs waiting.

**Deferred requirements:** LIB-3, LIB-8

---

## Cancel Resets Queued Folders to Downloaded

When the user presses `[c] cancel` during processing, remaining queued folders should have their status explicitly reset to `downloaded`. Currently the queued flag is cleared but the persisted `status.json` is not updated, leaving orphaned state.

**Deferred requirements:** LIB-15, PROC-15

---

## Delete Confirmation Dialog

`[d] delete` on the import screen immediately removes folders from the device with no confirmation. A destructive action of this kind should require an explicit confirmation step before executing (UX standards §4).

---

## Device Switching

When multiple device sources are configured in `config.json`, the import screen should allow switching between them at runtime. The device name should show a `▾` indicator and `[s] switch device` should open an inline selector. Selecting a device validates it before switching; an inaccessible device shows an inline error and leaves the current device active.

**Deferred requirements:** IMP-3, IMP-4, IMP-5, IMP-6, IMP-8 (partial), DEV-9

---

## Post-Processing Summary Screen

After all queued recordings finish transcribing, the app should show a summary screen listing what was processed, with the option to open any completed transcript directly. From the summary, the user can navigate to the library or back to import.

**Deferred requirements:** PROC-21, PROC-22, PROC-23

---

## Overall Queue Progress Indicator

During processing, the footer or header should show a "Recording X of Y" counter so the user knows how far through the queue they are. Currently only per-recording progress is shown.

**Deferred requirements:** PROC-14 (partial)

---

## Drop Folder / Watch Folder

Process `.opus` files placed manually into a folder without a connected device.

**Why deferred:** OQ-10. Requires deciding on folder structure, naming
conventions, and how to surface these files in the import/library flow.
Device-based import covers v1 use cases.
