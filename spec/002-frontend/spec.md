# Trustate Document Intake Assistant — Frontend Spec

Single HTML page. No framework, no build step, no dependencies beyond the
browser. Delivered as `static/index.html`, served by the Go backend.

---

## Stack

- Pure HTML + CSS + vanilla JS (ES2020)
- No npm, no bundler, no React
- One file: `static/index.html`
- Backend served on `http://localhost:8080`

---

## Go Backend Change (one addition)

Add `GET /` to `main.go` serving `static/index.html` as `text/html`.
Optionally use `http.FileServer` on `./static/` for simplicity.

---

## Layout (single page, two panels)

```
┌─────────────────────────────────────────────────┐
│  Trustate — Document Intake Assistant            │
├───────────────────┬─────────────────────────────┤
│  LEFT PANEL       │  RIGHT PANEL                │
│                   │                             │
│  [ Load Sample ▾] │  Extracted Facts            │
│  [ will ] [ poa ] │  ┌──────────────────────┐  │
│                   │  │ Label    │ Value       │  │
│  — or —           │  │ Testator │ …           │  │
│                   │  │ Executor │ …           │  │
│  [ Upload PDF ]   │  └──────────────────────┘  │
│  (drag or click)  │                             │
│                   │  Ask a question             │
│  Status / spinner │  ┌──────────────────────┐  │
│                   │  │ Type your question…  │  │
│                   │  └──────────────────────┘  │
│                   │  [ Ask ]                    │
│                   │                             │
│                   │  ── Answer ──               │
│                   │  (answer text here)         │
└───────────────────┴─────────────────────────────┘
```

---

## State Machine

```
IDLE → LOADING → READY → ASKING → READY
         ↓                  ↓
       ERROR              ERROR
```

- **IDLE**: no session. Upload area and sample buttons active. Right panel empty.
- **LOADING**: file sent to `/upload`. Spinner shown. Inputs disabled.
- **READY**: session exists. Facts table populated. Ask input + button active.
- **ASKING**: request in flight to `/ask`. Ask button disabled, spinner inline.
- **ERROR**: red status message. Upload area re-enabled.

---

## Components

### Header
```html
<header>
  <h1>Trustate — Document Intake Assistant</h1>
</header>
```
One line. No nav, no logo.

---

### Left Panel — Document Input

**Sample buttons**
```html
<button id="btn-will">Load Will Sample</button>
<button id="btn-poa">Load POA Sample</button>
```
On click: `GET /sample/{name}` → receive blob → treat as File → trigger upload
flow (same as manual upload). Buttons disabled during LOADING.

**Upload area**
```html
<div id="drop-zone">
  <p>Drag a PDF here or <label for="file-input">browse</label></p>
  <input type="file" id="file-input" accept=".pdf" hidden>
</div>
```
- Click on label opens file picker
- Dragover: highlight border
- Drop: accept first file if `.pdf`, reject otherwise (show error)
- On file selected (either path): call `uploadFile(file)`

**Status line**
```html
<p id="status"></p>
```
Shows: "Uploading…" / "Ready." / "Error: …" in plain text.
No color coding beyond red for errors.

---

### Right Panel — Facts Table

```html
<section id="facts-section" hidden>
  <h2>Extracted Facts</h2>
  <table id="facts-table">
    <thead><tr><th>Field</th><th>Value</th></tr></thead>
    <tbody></tbody>
  </table>
</section>
```
Hidden until first successful upload. Cleared and repopulated on each upload.

---

### Right Panel — Q&A

```html
<section id="qa-section" hidden>
  <h2>Ask a Question</h2>
  <input type="text" id="question-input"
         placeholder="e.g. Who is the executor?">
  <button id="ask-btn">Ask</button>

  <div id="answer-box" hidden>
    <h3>Answer</h3>
    <p id="answer-text"></p>
  </div>
</section>
```
Hidden until READY state. On Ask click (or Enter key in input): call `askQuestion()`.

---

## API Calls

### `uploadFile(file: File)`

```js
async function uploadFile(file) {
  setState('LOADING');
  const form = new FormData();
  form.append('file', file);
  const res = await fetch('/upload', { method: 'POST', body: form });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error);
  sessionId = data.session_id;
  renderFacts(data.facts);
  setState('READY');
}
```

### `askQuestion()`

```js
async function askQuestion() {
  setState('ASKING');
  const res = await fetch('/ask', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ session_id: sessionId, question: questionInput.value })
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error);
  renderAnswer(data.answer);
  setState('READY');
}
```

Both wrapped in `try/catch` → `setState('ERROR', err.message)`.

---

## CSS Requirements

- Two-column flex layout (left ~35%, right ~65%)
- Mobile: single column stack (media query `max-width: 640px`)
- Drop zone: dashed border, highlight on dragover
- Facts table: full width, alternating row background
- Spinner: CSS-only rotating border circle, shown during LOADING / ASKING
- No external fonts, no external CSS
- Color palette: neutral grays + one accent (`#1a5276` — conservative, legal-appropriate)

---

## Frontend Task Breakdown

### Phase F0 — Static Serving

- [ ] **F001** — Add `./static/` directory; register `http.FileServer` on
  `GET /` in `main.go` serving `./static/`. _(deps: T012)_

### Phase F1 — HTML Skeleton

- [ ] **F010** — `static/index.html`: `<!DOCTYPE html>`, `<head>` (charset,
  viewport, title, inline `<style>` placeholder), `<body>` with header +
  two-panel flex container, all element IDs defined, sections hidden.
  _(deps: F001)_
- [ ] **F011** — Verify: `go run .` → `http://localhost:8080` renders blank
  two-panel layout with header. _(deps: F010)_

### Phase F2 — CSS

- [ ] **F020** — Inline styles in `<head>`: layout (flex, column widths),
  drop zone (dashed border, dragover class), facts table (full width, striped),
  spinner (CSS animation), status line (red for `.error` class), responsive
  breakpoint. _(deps: F010)_

### Phase F3 — Upload Flow

- [ ] **F030** — JS: `setState(state, msg?)` — toggle element visibility and
  `disabled` attributes per state machine above. _(deps: F010)_
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
  appears with populated rows. _(deps: F033, T043)_

### Phase F4 — Q&A Flow

- [ ] **F040** — JS: `askQuestion()` — JSON POST to `/ask`, `renderAnswer`,
  error handling. Enter key on `#question-input` triggers same function.
  _(deps: F032)_
- [ ] **F041** — Verify: type "Who is the executor?" → answer appears, no
  `{{…}}` tokens visible. _(deps: F040, T071)_

### Phase F5 — README Update

- [ ] **F050** — Append to `README.md`: one line noting the UI is served at
  `http://localhost:8080` after `go run .`. _(deps: T080, F041)_

---

## Dependency Summary

```
T012 → F001 → F010 → F011
                └─ F020
                └─ F030 → F031 → F032 → F033 → F034
                                    └─ F040 → F041
F041, T080 → F050
```

---

## Out of Scope

- Conversation history / multi-turn chat UI
- File name display or document preview
- Markdown rendering in answers
- Loading skeleton / progressive fact reveal
- Any form of authentication UI
- Mobile-specific touch gestures beyond responsive layout