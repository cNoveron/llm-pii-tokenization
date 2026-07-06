# Spec 005: Automation Tab Drag & Drop Interface

This specification defines the redesign of the **Automate** tab page to support dragging stages from a components sidebar into the **Actions Panel** for execution.

---

## Layout Redesign

The layout of the `#automate-page` is a two-column flex layout (sidebar on the left, Actions Panel on the right). The previous top container ("Automation Pipeline") is removed entirely.

```
┌────────────────────────────────────────────────────────┐
│  Trustate — Document Intake Assistant                  │
├────────────────────────────────────────────────────────┤
│  Scan  [ AUTOMATE ]                                    │
├────────────────────────────────────────────────────────┤
│  ┌──────────────────────────┬───────────────────────┐  │
│  │ SIDEBAR (STAGES)         │ ACTIONS PANEL         │  │
│  │                          │                       │  │
│  │  [Stage 1: Discovery ✥]  │  AI Provider: [o] V   │  │
│  │  [Stage 2: Retrieval ✥]  │  ┌─────────────────┐  │  │
│  │  [Stage 3: Actions ✥]    │  │ Drop stage here │  │  │
│  │                          │  └─────────────────┘  │  │
│  │                          │  (Actions results     │  │
│  │                          │   thread)             │  │
│  └──────────────────────────┴───────────────────────┘  │
└────────────────────────────────────────────────────────┘
```

---

## Components

### 1. Drag & Drop Sidebar (`#automate-sidebar`)
Located on the left (~25% width). It displays the available stages as draggable component cards.
- **Draggable elements** must have `draggable="true"`.
- Draggable stages:
  - **Stage 1 — Asset Discovery**: Scans document context for assets, bank accounts, or properties.
  - **Stage 2 — Data Retrieval**: Simulates gathering records/verification data.
  - **Stage 3 — Legal Actions**: Generates legal instructions or draft forms based on the document.

### 2. Actions Panel (`#actions-panel`)
Located on the right (~75% width).
- Contains the AI Provider selector.
- Contains the **Drop Zone** (`#automate-drop-zone`). When empty, it displays instructions: *"Drag and drop a pipeline stage component here to execute it."*
- Contains the **Actions Thread** (`#actions-thread`), which displays the executing and completed stages as cards with status indicator, spinner, and results.

---

## Interaction and Behavior

### Drag & Drop Protocol
1. **Drag Start (`dragstart`)**:
   - The dragged stage element sets data payload: `e.dataTransfer.setData('text/plain', stageId)`.
   - The dragged element is visually dimmed (e.g., lower opacity).
2. **Drag Over / Enter (`dragover`, `dragenter`)**:
   - The drop zone (`#automate-drop-zone`) highlights its border/background to indicate it is a valid target.
3. **Drag Leave / End (`dragleave`, `dragend`)**:
   - Remove drop zone highlight.
   - Restore opacity of the dragged element.
4. **Drop (`drop`)**:
   - Remove drop zone highlight.
   - Extract `stageId` from the transfer payload.
   - Render a new active stage card in the Actions Panel.
   - Trigger the corresponding backend call:
     - Stage 1: `POST /automate/discover`
     - Stage 2: `POST /automate/retrieve`
     - Stage 3: `POST /automate/actions`

---

## Implementation Details

### HTML Structural Changes
- Remove `.automate-section` containing `<h3>Automation Pipeline</h3>`.
- Insert two-column layout:
  ```html
  <div id="automate-page" class="hidden">
    <div id="automate-container">
      <aside id="automate-sidebar">
        <h3>Pipeline Components</h3>
        <div class="draggable-stage" draggable="true" data-stage="1">
          <h4>Stage 1 — Asset Discovery</h4>
          <p>Scan document context for estate assets.</p>
        </div>
        <div class="draggable-stage" draggable="true" data-stage="2">
          <h4>Stage 2 — Data Retrieval</h4>
          <p>Simulate document/data retrieval.</p>
        </div>
        <div class="draggable-stage" draggable="true" data-stage="3">
          <h4>Stage 3 — Legal Actions</h4>
          <p>Generate draft actions & estate plans.</p>
        </div>
      </aside>

      <main id="actions-panel" class="automate-section">
        <h3>Actions Panel</h3>
        <!-- AI Provider Selector -->
        ...
        <!-- Drop Zone -->
        <div id="automate-drop-zone">
          <p>Drag and drop a pipeline stage component here to execute it</p>
        </div>

        <div id="actions-thread"></div>
      </main>
    </div>
  </div>
  ```

### CSS Requirements
- `#automate-container`: flex or grid container for sidebar and panel.
- `.draggable-stage`: cursor pointer or move, border, light shadow, and draggable styling.
- `#automate-drop-zone`: dashed border, light background, centered instruction text. Highlights on `.dragover`.
- Responsive behavior: stack vertically on narrow viewports.
