# Trustate Document Intake Assistant — Frontend Implementation Plan

Derived from [`spec.md`](./spec.md). This plan translates the specification into an ordered, phase-by-phase implementation strategy.

---

## 1. Architecture Overview

A single-page HTML application served statically by the Go backend. It uses vanilla HTML, CSS, and JS (ES2020) without any build step, dependencies, or frameworks. The frontend manages its own state and communicates with the backend via REST API endpoints.

```
                    ┌─────────────────────────────────────────┐
                    │  static/index.html                      │
                    │    ├─ HTML Skeleton & Layout            │
                    │    ├─ CSS Styling (Inline)              │
                    │    └─ Vanilla JS State & API Logic      │
                    └───────────────────┬─────────────────────┘
                                        │ (Fetch API)
                                        ▼
                    ┌─────────────────────────────────────────┐
                    │  Go Backend (http://localhost:8080)     │
                    │    ├─ GET /sample/{name}                │
                    │    ├─ POST /upload                      │
                    │    └─ POST /ask                         │
                    └─────────────────────────────────────────┘
```

**State Machine:**
The UI transitions between IDLE, LOADING, READY, ASKING, and ERROR states, controlling visibility and interactions of the Left (Input) and Right (Results/Q&A) panels.

---

## 2. Tech Stack Decisions

| Concern | Decision | Rationale |
|---|---|---|
| Hosting | Served via Go `http.FileServer` | Meets spec requirement of zero-dependency static serving. |
| Structure | Single HTML file (`index.html`) | Simplifies deployment and development, no bundler needed. |
| Styling | Vanilla CSS in `<style>` block | Meets spec; flexbox provides two-column layout and mobile responsiveness. |
| Logic | Vanilla JavaScript in `<script>` block | Meets ES2020 spec; uses `async/await`, `fetch`, and DOM manipulation. |
| File Uploads | `FormData` API via `POST /upload` | Standard, browser-native multipart uploads. |
| API Communication | `fetch` API | Standard native HTTP client. |
| Error Handling | Catch-all `try/catch` wrapping API calls | Consistently routes errors to the UI's ERROR state. |

---

## 3. Package & File Plan

All frontend assets are contained within `static/index.html`, which will be served by the Go server.
Changes to the Go backend `main.go` will be minimal, specifically to mount the static file server.

| File | Responsibility |
|---|---|
| `main.go` | (Backend) Add `http.FileServer` routing for `GET /` to `./static/` |
| `static/index.html` | Entire frontend application (HTML, CSS, JS) |
| `README.md` | Update to mention static serving at `http://localhost:8080` |

---

## 4. Design Detail: CSS & Layout

- **Layout Grid**: CSS Flexbox on the main container. Left panel is roughly `35%`, Right panel is roughly `65%`.
- **Responsive Design**: `@media (max-width: 640px)` changes the flex direction to `column` so the panels stack.
- **Drop Zone**: Styled with dashed borders. Uses JavaScript `dragover`, `dragleave`, and `drop` events to toggle a dragover highlight class.
- **Spinner**: A pure CSS rotating border circle, toggled by the state machine during LOADING and ASKING.
- **Color Palette**: Neutral grays with `#1a5276` as the primary accent color (conservative, legal-appropriate).
- **Typography**: Uses system fonts. No external fonts.

---

## 5. Design Detail: State Management & JavaScript

A central `setState(state, msg)` function governs the UI.

- **IDLE**: No session. Upload area and sample buttons active. Right panel is empty.
- **LOADING**: File sent to `/upload`. Shows spinner. Disables sample buttons and upload inputs.
- **READY**: Session exists. Facts table populated and Q&A section shown. Ask input + button active.
- **ASKING**: Request in flight to `/ask`. Disable ask button, show spinner inline.
- **ERROR**: Re-enables upload area, displays red error message `msg`.

State transitions trigger specific DOM updates (adding/removing the `hidden` attribute or `disabled` props).

---

## 6. Milestones (build & verify order)

Following the spec's dependency graph:

1. **Phase F0: Static Serving** — Add `./static/` directory; register `http.FileServer` on `GET /` in `main.go`.
2. **Phase F1: HTML Skeleton** — Create `static/index.html` with basic elements, IDs, and hidden sections. Verify `http://localhost:8080` serves the file.
3. **Phase F2: CSS** — Add the inline `<style>` block. Apply flexbox layout, dashed drop zone, facts table alternating rows, and the CSS spinner.
4. **Phase F3: Upload Flow** — Implement `setState` and file upload handling (`change`, `drop` events). Call `POST /upload`. Populate the facts table. Wire up the sample buttons to trigger `GET /sample/{name}` and start upload. Verify end-to-end upload with samples.
5. **Phase F4: Q&A Flow** — Implement `POST /ask` on "Ask" button click or Enter key. Render response in the UI.
6. **Phase F5: README Update** — Update `README.md` to mention the UI link.

---

## 7. Testing & Verification Strategy

- **Manual UI Check:** Verify two-column layout on desktop, single-column stack below 640px.
- **Drag & Drop:** Verify the drop zone highlights correctly and rejects non-PDFs gracefully with the correct error state.
- **Sample Button Flow:** Click a sample button and ensure the file is retrieved, uploaded, the Facts table appears, and the Q&A section becomes active.
- **Q&A Flow:** Submit a question, ensure the spinner displays, the answer is rendered, and no tokens (like `{{PERSON_1}}`) are visible in the answer text.
- **Error Handling:** Simulate an error (e.g., upload a bad file) and ensure the UI transitions to ERROR and remains usable.

---

## 8. Out of Scope

- Chat history or multi-turn conversational UI.
- Document preview or showing the selected file's name.
- Markdown rendering in LLM answers.
- Loading skeletons or progressive fact reveal.
- Any form of authentication UI.
- Mobile-specific touch gestures beyond standard responsive layout.

---

## 9. Risks & Mitigations

| Risk | Mitigation |
|---|---|
| Fetch API CORS issues on sample load | The frontend is served on the same origin (`http://localhost:8080`), bypassing CORS issues. |
| State desynchronization | All UI changes are funneled through the central `setState` function to ensure consistency. |
| Incomplete backend | Ensure backend implementation plan milestones (M1-M7) are completed before starting Phase F3/F4 frontend integration. |
