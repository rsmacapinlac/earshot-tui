# Import Screen Requirements

## Overview

The import screen presents recordings from a connected earshot device,
organised by folder. The user selects folders to import and downloads them
to local storage.

## Entry Points

- App launch when device is accessible (after preflight and config checks)
- `[i] import` from the library screen

## Exit Points

- `[l] library` → library screen
- `[q] quit` → exit app

## Layout

```
  Device: Pi4-Earshot ▾

  [✓] 2026-03-31 09:14/    3 recordings   8m 45s
  [ ] 2026-03-30 14:22/    2 recordings  14m 22s
  [✓] 2026-03-29 08:55/    1 recording    3m 12s

  ──────────────────────────────────────────────────────
  [space] select   [i] import   [l] library   [q] quit
```

During import:

```
  Device: Pi4-Earshot ▾

  [↓] 2026-03-31 09:14/    Downloading... ████████░░  82%
  [ ] 2026-03-30 14:22/    2 recordings  14m 22s
  [✓] 2026-03-29 08:55/    1 recording    3m 12s

  ──────────────────────────────────────────────────────
  [c] cancel
```

Scanning device:

```
  Device: Pi4-Earshot ▾

  Scanning...

  ──────────────────────────────────────────────────────
  [l] library   [q] quit
```

Device not accessible:

```
  Device: Pi4-Earshot ▾

  Pi4-Earshot not found at /run/media/ritchie/EARSHOT. Is it mounted?

  ──────────────────────────────────────────────────────
  [r] retry   [s] switch device   [l] library   [q] quit
```

## Device Selector

- IMP-1: On launch, the import screen defaults to the first device source in
  `config.json`. Its name is displayed at the top of the screen.
- IMP-2: If only one device source is registered, no switcher indicator is
  shown.
- IMP-3: If more than one device source is registered, the device name is
  followed by `▾` and `[s] switch device` appears in the footer.
- IMP-4: Pressing `[s]` opens an inline select listing all registered device
  sources. Selecting one validates the device before switching.
- IMP-5: If the selected device is not accessible (not mounted or path
  invalid), an inline error is shown and the current device remains active:
  `"[device name] not found at [path]. Is it mounted?"`
- IMP-6: Device switching is hidden during an active download.

## Scanning

- IMP-7: On entry, the screen shows a `Scanning...` state while reading the
  device for folders. Navigation is available during scanning (`[l]`, `[q]`).
- IMP-8: If the device is not accessible on entry, an inline error is shown
  with `[r] retry`, `[s] switch device`, `[l] library`, and `[q] quit`.
- IMP-9: `[r] retry` re-runs the device check and scan without leaving the
  screen.

## Folder List

- IMP-10: Recordings are presented grouped by folder as they appear on the
  device. Each row shows: folder name (with timestamp derived from the folder
  name), recording count, and total duration summed from the `.opus` files
  inside the folder.
- IMP-11: The folder is the unit of selection — the user imports all recordings
  in a folder or none. There is no per-file selection.
- IMP-12: All folders default to unselected on entry.
- IMP-13: `[↑]` / `[↓]` navigates between folders. `[space]` toggles the
  focused folder.
- IMP-14: Folders are sorted most recent first.
- IMP-15: If no new recordings are found, the screen reads:
  "No new recordings on [device name]." Only `[l] library` and `[q] quit`
  are shown in the footer.

## Import

- IMP-16: `[i] import` begins downloading all selected folders in list order.
- IMP-17: `[i] import` is only shown when at least one folder is selected and
  no download is in progress.
- IMP-18: During download, the active folder row shows an inline progress bar.
  The checkbox is replaced by `[↓]`.
- IMP-19: When a folder finishes downloading it is removed from the list. The
  next selected folder begins immediately.
- IMP-20: When all selected downloads complete, the list shows only remaining
  un-imported folders. The user may select more and import again.
- IMP-21: A folder is only removed from the list once all its files are fully
  written to local storage. Partial downloads are discarded on cancel and the
  folder returns to unselected state (see engineering-principles.md §3).

## Cancellation

- IMP-22: `[c] cancel` is only shown during an active download.
- IMP-23: On cancel, the in-progress download stops, no partial files are
  retained, and any remaining queued folders return to unselected state.

## Navigation

- IMP-24: `[l] library` and `[q] quit` are hidden during an active download.
- IMP-25: `[l] library` and `[q] quit` are available at all other times.
