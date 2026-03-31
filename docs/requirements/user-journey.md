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
[Connect screen] ← PRIMARY entry point
    │
    ├── One device registered → auto-select
    ├── Multiple devices registered → pick from list
    └── No devices registered → guided first-time setup
                │
                └── Auto-scan common mount paths (/media/, /Volumes/, etc.)
                    Detect earshot device → propose hostname as device name
                    User confirms or edits name → device registered
    │
    ▼
[Scan device]
    │
    ├── New recordings found → [Disposition screen]
    │       │
    │       │  Act on each recording in priority order:
    │       ├── [d] download → queued for processing (in chosen order)
    │       ├── [s] skip     → ignored this session
    │       ├── [X] delete   → confirm (Huh inline) → removed from device + device DB updated
    │       ├── [a] download all → queues all remaining
    │       └── [alt+↑/↓] reorder → change processing priority
    │               │
    │               ▼
    │           [Processing] ← auto-starts after download
    │               │  Progress bar per recording, overall queue progress
    │               │  [C] Cancel available throughout
    │               │
    │               └── Complete → show processed list → select to open in $EDITOR
    │
    └── No new recordings → "All caught up." → [Library] (secondary)
```

## Secondary Journey: Revisit a Recording

The user comes back to a previously processed recording to listen and
identify speakers.

```
[Library view]
    │
    └── Select recording
            │
            ▼
        [Recording detail]
            │
            ├── [P] Play audio inline
            ├── [O] Open transcript in $EDITOR
            ├── [R] Rename speakers  ← only shown if 2 speakers detected
            │       │
            │       ├── Speaker 1: [____________]
            │       └── Speaker 2: [____________]
            │               │
            │               └── Saves → rewrites .md in place
            │
            └── [D] Delete local audio  ← to save disk space, confirm first
```

## Screen Inventory

| Screen             | Entry point                              | Exit                        |
|--------------------|------------------------------------------|-----------------------------|
| Connect            | App launch                               | Device selected             |
| First-time setup   | No registered devices                    | Device registered → Connect |
| Device list        | Multiple registered devices              | Device selected             |
| Disposition        | New recordings found after scan          | All acted on → Processing   |
| Processing         | Download confirmed                       | Complete / Cancelled        |
| Library            | No new recordings, or from any screen    | Recording detail / Connect  |
| Recording detail   | Select recording in library              | Library                     |
| Rename speakers    | [R] on recording detail (2 speakers only)| Recording detail            |
