# User Journey

## Primary User

The owner of one or more earshot devices. They record conversations in the
field and return to their desk to process them. They launch this app
specifically when they have a device to process — it is a workflow tool,
not a background service.

## End-to-End Journey

```
[Launch app]
    │
    ▼
[Preflight check] ← silent if all pass; errors exit with actionable message
    │
    ▼
[Config check]
    │
    ├── config.json valid → continue
    └── config.json missing or incomplete → [Setup wizard]
            │
            └── Device source (auto-scan or manual entry)
                    │
                    └── Config written → continue
    │
    ▼
[Import screen] ← device accessibility checked here inline
    │
    ├── [space] toggle folder selection
    ├── [i] import → downloads selected folders (progress inline)
    ├── [c] cancel → stops active download
    ├── [s] switch device → validates device before switching
    └── [l] library → [Library]
    │
    ▼
[Library] ← main screen
    │
    ├── [space] select downloaded/failed/interrupted folders
    ├── [t] transcribe → transcription runs inline
    ├── [enter] on transcribed folder → opens transcript in $EDITOR
    ├── [b] back → [Import screen]
    └── [q] quit
```

## Screen Inventory

| Screen       | Entry point                                   | Exit                        |
|--------------|-----------------------------------------------|-----------------------------|
| Setup wizard | config.json missing or incomplete             | Config written → Import     |
| Import       | App launch (after preflight), or [b] from Library | [l] library → Library   |
| Library      | [l] from Import                               | [b] back → Import / [q] quit |
