# Setup Wizard Requirements

## Overview

The setup wizard runs when `config.json` is missing or `device_sources` is
absent or empty. It collects the minimum configuration required for the app
to function — one registered device source — and writes `config.json` before
proceeding to the import screen.

The wizard is not a one-time experience — it runs whenever the config is
invalid, including after a user manually edits `config.json` incorrectly.

## Entry Points

- `config.json` does not exist
- `config.json` exists but `device_sources` is empty or absent

## Exit Points

- Config written successfully → import screen

## Flow

```
[Setup wizard]
    │
    └── Device source
            Auto-scan common mount paths for earshot devices:
              Linux:  /media/$USER/, /mnt/
              macOS:  /Volumes/
            │
            ├── Device found → propose hostname as name, show mount path
            │       User confirms or edits name → source added
            │
            └── No device found → prompt for mount path manually
                    User enters path and name → source added
    │
    ▼
[Write config.json] → [Import screen]
```

## Layout

Device found:

```
  Welcome to earshot-tui.

  Earshot device found.

  Name:  [Pi4-Earshot   ]
  Path:   /run/media/ritchie/EARSHOT

  ──────────────────────────────────────────────────────
  [enter] confirm   [e] edit name
```

No device found:

```
  Welcome to earshot-tui.

  No earshot device detected.

  Name: [                  ]
  Path: [                  ]

  ──────────────────────────────────────────────────────
  [enter] confirm
```

## Requirements

### Device Source

- SETUP-1: On entry, the app auto-scans common mount paths for a directory
  containing an earshot device structure (`earshot.db` and `recordings/`).
- SETUP-2: If a device is found, its hostname is proposed as the device name.
  The mount path is shown read-only. The user may edit the name before confirming.
- SETUP-3: If no device is found, the user is prompted to enter both the device
  name and mount path manually.
- SETUP-4: On confirm, the mount path is validated as accessible. If not, an
  inline error is shown and the user is re-prompted.
- SETUP-5: Only one device source is collected during setup. Additional sources
  can be added by editing `config.json` directly.

### Config Write

- SETUP-6: On confirming the device source, `config.json` is written atomically
  (write to temp, rename) to `AppDirs.Config()/config.json`.
- SETUP-7: If writing fails, a specific error is shown with the path and
  reason. The wizard does not proceed.
- SETUP-8: On success, the app proceeds directly to the import screen
  without restarting.
