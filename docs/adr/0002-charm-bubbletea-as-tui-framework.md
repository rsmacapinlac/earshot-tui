# ADR-0002: Charm (Bubble Tea) as TUI Framework

**Status:** Accepted

## Context

earshot-tui has a well-defined set of distinct screens with clear transitions
between them (connect → disposition → processing → library → recording detail).
The UX standards require keyboard-first navigation, a persistent footer showing
available actions, progress bars for long-running operations, and a UI that
remains responsive during background processing (see ../ux-standards.md).

Candidates considered:

- **bubbletea (Charm)**: Elm-architecture model (Model/Update/View). Screen
  transitions map naturally to model state changes. First-class async command
  support via `tea.Cmd` handles background work without blocking the UI.
  Actively maintained with a large ecosystem (Bubbles components, Lip Gloss
  styling, Huh forms).
- **tview**: Immediate-mode widget library. Good for dashboards but less
  suited to workflow-style screen transitions. Harder to reason about state.
- **tcell / raw curses**: Too low-level. Would require building component
  abstractions from scratch with no benefit for this use case.

## Decision

Use **Bubble Tea** as the TUI framework, with **Lip Gloss** for styling and
**Bubbles** for common components (progress bars, spinners, tables, text inputs).

## Consequences

- Each screen in the screen inventory (user-journey.md) maps to a Bubble Tea
  `Model` with its own `Update` and `View` functions. Transitions are handled
  by returning a new model from `Update`.
- Background processing (transcription, diarization) runs via `tea.Cmd`
  goroutines. Progress updates are sent back to the model as `tea.Msg` values —
  this satisfies NFR-3 (non-blocking UI) without custom concurrency management.
- Cancellation (PROC-15) is implemented by sending a cancel `tea.Msg` that
  signals the background command to stop.
- Lip Gloss enforces consistent colour usage across screens (../ux-standards.md §10).
- The Huh library handles form inputs (device naming, speaker renaming) with
  built-in validation and keyboard navigation.
- Bubble Tea's architecture makes the app straightforwardly testable — models
  are pure functions of state.
