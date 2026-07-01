# Trustate Document Intake Assistant — Frontend Tasks

Ordered task breakdown derived from [`plan.md`](./plan.md) and [`spec.md`](./spec.md). Tasks are grouped
by phase (following the plan's milestones). `[P]` = parallelizable with other
`[P]` tasks in the same phase. Each task lists its dependencies by ID.

Legend: `[ ]` pending · `[~]` in progress · `[x]` done

---

## Phase F0 — Static Serving

- [ ] **F001** — Add `./static/` directory; register `http.FileServer` on
  `GET /` in `main.go` serving `./static/`. _(deps: Backend T012)_

## Phase F1 — HTML Skeleton

- [ ] **F010** — `static/index.html`: `<!DOCTYPE html>`, `<head>` (charset,
  viewport, title, inline `<style>` placeholder), `<body>` with header +
  two-panel flex container, all element IDs defined, sections hidden.
  _(deps: F001)_
- [ ] **F011** — Verify: `go run .` → `http://localhost:8080` renders blank
  two-panel layout with header. _(deps: F010)_

## Phase F2 — CSS

- [ ] **F020** — Inline styles in `<head>`: layout (flex, column widths),
  drop zone (dashed border, dragover class), facts table (full width, striped),
  spinner (CSS animation), status line (red for `.error` class), responsive
  breakpoint. _(deps: F010)_ `[P]`

## Phase F3 — Upload Flow

- [ ] **F030** — JS: `setState(state, msg?)` — toggle element visibility and
  `disabled` attributes per state machine (`IDLE`, `LOADING`, `READY`, `ASKING`, `ERROR`). _(deps: F010)_
- [ ] **F031** — JS: file picker (`change` event on `#file-input`) and drag-
  and-drop (`dragover`, `drop` on `#drop-zone`) → call `uploadFile(file)`.
  Reject non-PDF with `setState('ERROR', 'Please upload a PDF file.')`.
  _(deps: F030)_
- [ ] **F032** — JS: `uploadFile(file)` — FormData POST to `/upload`,
  `renderFacts(facts)` (populate `<tbody>`), store `sessionId`, error handling.
  _(deps: F031)_
- [ ] **F033** — JS: sample buttons → `GET /sample/{name}` → `response.blob()`
  → `new File([blob], '{name}.pdf', {type:'application/pdf'})` → `uploadFile`.
  _(deps: F032)_
- [ ] **F034** — Verify: click "Load Will Sample" → spinner → facts table
  appears with populated rows. _(deps: F033, Backend T043, Backend T062)_

## Phase F4 — Q&A Flow

- [ ] **F040** — JS: `askQuestion()` — JSON POST to `/ask`, `renderAnswer`,
  error handling. Enter key on `#question-input` triggers same function.
  _(deps: F032)_
- [ ] **F041** — Verify: type "Who is the executor?" → answer appears, no
  `{{…}}` tokens visible. _(deps: F040, Backend T071)_

## Phase F5 — README Update

- [ ] **F050** — Append to `README.md`: one line noting the UI is served at
  `http://localhost:8080` after `go run .`. _(deps: F041, Backend T080)_

---

## Dependency Summary

```text
Backend T012 → F001 → F010 → F011
                        └─ F020
                        └─ F030 → F031 → F032 → F033 → F034
                                            └─ F040 → F041
F041, Backend T080 → F050
```

## Coverage Check

| Spec area | Task(s) |
|---|---|
| Single HTML page / Static Serving | F001, F010 |
| Flex Layout / Responsive CSS | F020 |
| State Machine (IDLE/LOADING/READY/ASKING/ERROR) | F030 |
| Left Panel (Drop Zone / Manual Upload) | F031, F032 |
| Left Panel (Sample Buttons) | F033, F034 |
| Right Panel (Facts Table Rendering) | F032 |
| Right Panel (Q&A Flow / Answer Rendering) | F040, F041 |
| README Update | F050 |
