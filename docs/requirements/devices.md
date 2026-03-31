# Device Management Requirements

## Overview

An earshot device is a Raspberry Pi running the earshot application. The TUI
supports one or more registered devices. Connection is via USB, manually
initiated by the user.

## Requirements

### First-Time Setup

- DEV-1: If no devices are registered on launch, the app enters a guided
  setup flow automatically — it does not present an empty screen.
- DEV-2: The app scans common USB mount paths for a recognisable earshot
  device structure (presence of `earshot.db` and `recordings/` directory):
  - macOS: `/Volumes/`
  - Linux: `/media/$USER/`, `/mnt/`
- DEV-3: If a device is found, the app proposes the device's hostname
  (read from the mounted filesystem) as the device name. The user may
  edit it before confirming.
- DEV-4: If no device is auto-detected, the user is prompted to enter the
  mount path manually.

### Registration

- DEV-5: Registered devices are stored in the local application database.
- DEV-6: The user can list all registered devices.
- DEV-7: The user can edit a device's name or mount path.
- DEV-8: The user can remove a registered device. This does not affect the
  device itself.

### Connection

- DEV-9: Connecting a device is manually initiated — the user selects a
  registered device and confirms it is mounted.
- DEV-10: If only one device is registered, it is auto-selected on launch.
- DEV-11: The app verifies the mount path is accessible before proceeding.
  If not found, display a specific error with the path and a suggestion to
  check the mount (see ../ux-standards.md §6).
- DEV-12: The app reads the earshot SQLite database (`earshot.db`) on the
  mounted device to determine recording state.
- DEV-13: Only one device is active per session.

### Device Disconnection

- DEV-14: If the device is disconnected mid-session, the app handles this
  gracefully: pause any active download, display a clear error, return to
  the connect screen. No crash.
