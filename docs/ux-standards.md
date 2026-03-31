# UX Standards

Principles and interaction patterns that guide all UI decisions in earshot-tui.
When in doubt, refer here before adding new interactions or screens.

The TUI framework is **Bubble Tea** (Charm). Where a standard references a
specific implementation pattern, it is written against Bubble Tea's model.

---

## 1. Keyboard-First, Always

This is a TUI. Every action must be reachable via keyboard. Mouse support is
not required and should not be designed for. Navigation must feel natural to
users familiar with terminal applications.

- `↑` / `↓` or `j` / `k` for list navigation
- `enter` to select / confirm
- `esc` to go back or cancel
- Single-letter shortcuts for primary actions — **always lowercase** unless
  capital is intentional friction (see §4)
- `q` to quit from any top-level screen that is not actively processing

### Key naming convention

Keys are shown in the footer as `[d]ownload`, `[s]kip`, `[x] delete`. The
bracket wraps the key to press. Lowercase means the key is pressed without
modifier. Uppercase (e.g., `[X]`) is reserved for destructive actions where
Shift acts as a deliberate speed-bump.

---

## 2. Always Show Available Actions

The user should never wonder "what can I do here?" Every screen must display
available keybindings using Bubble Tea's **Bubbles `help` component** rendered
as a footer bar at the bottom of the screen.

The help component renders in two modes, toggled with `?`:

- **Short mode** (default): one line, most-used keys only
- **Long mode**: expanded, all available keys grouped by category

```
Short:  [↑↓] navigate   [d] download   [s] skip   [?] more
Long:   navigation: [↑↓] move   [enter] select
        actions:    [d] download   [s] skip   [X] delete   [a] all
        app:        [esc] back   [q] quit   [?] hide help
```

Do not show actions that are not available in the current context. The Bubbles
help component supports context-sensitive binding lists — use this. Do not
render unavailable keys as greyed-out; omit them entirely.

---

## 3. Context-Sensitive UI

Adapt the interface to what is actually relevant. Do not present options the
user cannot or should not use. Hide, do not disable:

- `[r] rename` → only shown when 2 speakers are detected
- `[d] delete audio` → only shown after processing is complete and local audio exists
- `[c] cancel` → only shown when processing is actively running
- `[s] stop` → only shown when audio is actively playing

Hiding irrelevant options reduces cognitive load and prevents errors. In Bubble
Tea, implement this by conditionally including or excluding bindings from the
help model and conditionally handling keys in `Update()`.

---

## 4. Destructive Actions Require Confirmation

Any action that cannot be undone requires an explicit confirmation step.

### Key friction

Destructive actions use **`[X]` (Shift+x)** rather than `[x]`. The Shift
modifier is intentional friction — it makes accidental triggering meaningfully
harder without requiring a multi-step flow.

### Confirmation implementation

Bubble Tea has no native modal system. Confirmations are implemented as
**inline state changes** using the Huh library's `Confirm` component,
rendering within the current screen, not as a floating overlay.

```
  Delete "2026-03-31 09:14" from device? This cannot be undone.
  > No   Yes
```

Default selection is always the safe option (No). The user must actively
move to Yes and press `enter`.

---

## 5. Long-Running Operations Are Non-Blocking

Transcription and diarization can take minutes. The UI must remain responsive
and navigable throughout.

- Show a progress bar (Bubbles `progress` component) per recording
- Show overall queue progress and elapsed time
- `[c]` cancels at any time — marks current recording as `interrupted`,
  returns remaining queued recordings to `downloaded` state
- Processing runs via `tea.Cmd` goroutines sending `tea.Msg` progress updates
  back to the model — the UI event loop is never blocked

### Quitting during processing

`q` during active processing does not quit immediately. Show an inline
confirmation (Huh `Confirm`):

```
  Processing in progress. Quit anyway? Recordings will resume on next launch.
  > No   Yes
```

If confirmed, the app exits. Recordings in `processing` are marked
`interrupted` on next launch (PROC-20).

### Long queues

A processing queue that exceeds the terminal height must scroll. Use the
Bubbles `viewport` component to wrap the queue list. The currently processing
item is always kept in view.

---

## 6. Errors Are Specific and Actionable

Do not show generic error messages. Tell the user what happened and what
they can do next:

| Situation | Bad | Good |
|---|---|---|
| Device not found | "Error" | "Device not found at /media/ritchie/earshot. Is it mounted?" |
| Transcription failed | "Processing failed" | "Transcription failed for 2026-03-31 09:14. [r] retry   [i] ignore" |
| Device disconnected | "Error" | "Device disconnected. Downloads paused. Reconnect to continue." |

---

## 7. External Tools: Two Distinct Patterns

The TUI delegates to external tools for editing and audio. These are **not
the same pattern** and must not be implemented the same way.

### Transcript editor — TUI suspends

Use `tea.ExecProcess()`. Bubble Tea suspends, hands full terminal control to
`$EDITOR`, and automatically resumes when the editor exits. No "press any
key" prompt is needed — resumption is automatic.

Notify the user before handing off:
```
  Opening transcript in editor. Return here when done.
```

### Audio player — TUI stays live

Use a background `exec.Command`, **not** `tea.ExecProcess`. The TUI remains
fully interactive while audio plays. The user can navigate, read the screen,
and press `[s]top` to end playback. The `AudioPlayer` interface (ADR-0005)
owns the player process lifecycle.

While audio is playing, the footer updates to surface the stop action:
```
  [s] stop   [o] open transcript   [esc] back
```

---

## 8. Graceful Recovery

The user should never lose work due to an interruption:

- If the app is closed during processing, mark in-progress recordings as
  `interrupted` on next launch and offer to retry (PROC-20)
- If the device is disconnected mid-download, report cleanly; do not mark
  the partial download as complete
- On every launch, reconcile local DB state before presenting any screen

---

## 9. Terminal Resize Is Handled Everywhere

Bubble Tea sends `tea.WindowSizeMsg` when the terminal is resized. Every
screen must handle this message and re-render correctly:

- Lists truncate or scroll to fit available height
- Progress bars resize to fit available width
- Viewports update their dimensions
- No screen may assume a fixed terminal size

A screen that breaks on resize is a broken screen.

---

## 10. Minimal, Purposeful Chrome

Earshot-tui is a workflow tool, not a dashboard:

- No decorative borders, ASCII art, or logo banners
- No status information that is not actionable
- Single-purpose screens — one task per screen
- The help footer (§2) is the only persistent chrome

---

## 11. Colour Usage

Implemented via **Lip Gloss** styles. Define all colours as named style
variables — never use raw colour codes inline.

| Colour  | Meaning | Usage |
|---------|---------|-------|
| Green   | Success / processed / complete | Status badges |
| Yellow  | Pending / in-progress / needs attention | Status badges |
| Red     | Error / failed / destructive warning | Badges, confirmations |
| Blue    | Active / currently playing | Playing state |
| Default | Neutral state, navigation, labels | Everything else |

Never use colour as the only indicator — always pair with text. This ensures
the app is usable on monochrome terminals and by users with colour blindness.
