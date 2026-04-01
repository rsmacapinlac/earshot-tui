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

## Drop Folder / Watch Folder

Process `.opus` files placed manually into a folder without a connected device.

**Why deferred:** OQ-10. Requires deciding on folder structure, naming
conventions, and how to surface these files in the import/library flow.
Device-based import covers v1 use cases.
