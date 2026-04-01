# Device Management Requirements

## Overview

An earshot device is a Raspberry Pi running the earshot application. The TUI
supports one or more registered devices stored in `config.json`. There is no
dedicated connect screen — the app always opens on the import screen, defaulting
to the first configured device source.

## Requirements

### First-Time Setup

- DEV-1: If `config.json` is missing or incomplete on launch, the app enters
  the setup wizard automatically (see setup-wizard.md).
- DEV-2: The setup wizard scans common USB mount paths for a recognisable
  earshot device structure (presence of `earshot.db` and `recordings/`
  directory):
  - macOS: `/Volumes/`
  - Linux: `/media/$USER/`, `/mnt/`
- DEV-3: If a device is found, the app proposes the device's hostname
  (read from the mounted filesystem) as the device name. The user may
  edit it before confirming.
- DEV-4: If no device is auto-detected, the user is prompted to enter the
  mount path manually.

### Registration

- DEV-5: Registered devices are stored as entries in the `device_sources`
  map in `config.json`.
- DEV-6: Additional devices can be added by editing `config.json` directly.
  The setup wizard only registers one device.

### Connection

- DEV-7: On launch, the app verifies the first device source in `config.json`
  is accessible before proceeding to the import screen. If not found, a
  specific error is shown with the path and a suggestion to check the mount.
- DEV-8: The app reads the earshot device filesystem to discover recording
  folders.
- DEV-9: When switching device sources on the import screen, the selected
  device is validated before the source switches. If not accessible, an inline
  error is shown and the current source remains active (see import.md IMP-5).
- DEV-10: Only one device is active per session.

### Device Disconnection

- DEV-11: If the device is disconnected mid-download, the app pauses the
  active download, displays a clear error, and cancels the remaining queue.
  No crash. The import screen remains showing the error.
