# Trustate Frontend Addendum — Chat Interface with PII Visibility

Extends `trustate-frontend-spec.md`. Replaces the Q&A section of the right
panel with a proper chat interface. All changes are confined to
`static/index.html` unless noted.

---

## What Changes

The existing single-question input + answer box is replaced by:
- A scrollable chat thread showing the **full LLM interaction** as bubbles
- On every `/ask` call, the thread renders four consecutive blocks:
  1. System prompt bubble (assistant-side, grey, fixed per session)
  2. Document context bubble (assistant-side, truncated tokenized full text)
  3. User bubble (the question as typed)
  4. Assistant response bubble (tokenized answer, togglable)
- A global toggle switches all tokenized content ↔ detokenized across the thread

The `/ask` backend endpoint requires a small extension (see below).
The `/upload` backend endpoint also requires a small extension (see below).

---

## Required Backend Changes

### `handler/upload.go` — extend `UploadResponse`

```go
type UploadResponse struct {
    SessionID      string `json:"session_id"`
    Facts          []Fact `json:"facts"`
    SystemPrompt   string `json:"system_prompt"`    // NEW: Q&A system prompt text
    TokenizedDoc   string `json:"tokenized_doc"`    // NEW: full tokenized document text
}
```

Populate by returning the Q&A system prompt string (already a constant in
`handler/ask.go` — move it to a shared location, e.g. `handler/prompts.go`)
and `session.TokenText` from the stored session.

### `handler/ask.go` — extend `AskResponse`

```go
type AskResponse struct {
    Answer          string `json:"answer"`            // de-tokenized (existing)
    TokenizedPrompt string `json:"tokenized_prompt"`  // NEW: tokenized user question
    TokenizedAnswer string `json:"tokenized_answer"`  // NEW: raw LLM response before detokenize
}
```

Populate before detokenizing:
```go
tokenizedAnswer  := rawLLMResponse
detokenizedAnswer := pii.Detokenize(rawLLMResponse, combinedMap)
tokenizedPrompt  := tokenizedQuestion // already computed during Ask flow
```

### `handler/prompts.go` — new shared file

Extract the Q&A system prompt into a package-level constant so both
`upload.go` (to return it) and `ask.go` (to use it) share the same string.

```go
package handler

const QASystemPrompt = `You are a helpful assistant for legal staff reviewing estate documents.
The document uses privacy tokens like {{PERSON_1}} instead of real names.
Preserve tokens exactly as-is in your responses.
Answer based only on the document provided. Be concise.`
```

---

## Updated Right Panel Layout

```
┌─────────────────────────────────────────────────┐
│  Extracted Facts                                 │
│  ┌────────────┬──────────────────────────────┐  │
│  │ Field      │ Value                        │  │
│  └────────────┴──────────────────────────────┘  │
├─────────────────────────────────────────────────┤
│  Conversation          [🔒 Show real data]       │ ← global toggle
│ ┌─────────────────────────────────────────────┐ │
│ │                                             │ │
│ │  ┌── SYSTEM ───────────────────────────┐   │ │
│ │  │ You are a helpful assistant for     │   │ │  ← system prompt bubble
│ │  │ legal staff…                        │   │ │
│ │  └─────────────────────────────────────┘   │ │
│ │                                             │ │
│ │  ┌── DOCUMENT CONTEXT ─────────────────┐   │ │
│ │  │ LAST WILL AND TESTAMENT OF          │   │ │  ← tokenized doc (truncated)
│ │  │ {{PERSON_1}}… [truncated – full     │   │ │
│ │  │ document sent to LLM]               │   │ │
│ │  └─────────────────────────────────────┘   │ │
│ │                                             │ │
│ │                  Who is the executor?  [👤] │ │  ← user bubble
│ │                                             │ │
│ │  ┌── RESPONSE ─────────────────────────┐   │ │
│ │  │ The executor is {{PERSON_2}},        │   │ │  ← tokenized answer
│ │  │ residing at {{ADDRESS_1}}.           │   │ │
│ │  └─────────────────────────────────────┘   │ │
│ │                                             │ │
│ └─────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────┐ [Send]     │
│  │ Ask a question…                 │            │
│  └─────────────────────────────────┘            │
└─────────────────────────────────────────────────┘
```

---

## Session Initialization — Thread Seeding

On successful `/upload` response, before the user asks anything, seed the
chat thread with two non-removable bubbles:

### 1. System prompt bubble

```html
<div class="message system">
  <span class="block-label">System Prompt</span>
  <p class="message-text">{systemPrompt}</p>
</div>
```

Static text — no tokenized/detokenized duality. Never hidden by the toggle.
Styled distinctly (light grey background, dashed left border).

### 2. Document context bubble

```html
<div class="message context">
  <span class="block-label">Document Context</span>
  <p class="tokenized">{truncatedTokenizedDoc}<br>
    <em class="truncation-notice">— truncated · full document sent to LLM —</em>
  </p>
  <p class="detokenized hidden">{truncatedRawDoc}<br>
    <em class="truncation-notice">— truncated · full document sent to LLM —</em>
  </p>
</div>
```

`truncatedTokenizedDoc` = first 400 characters of `tokenized_doc` from
`/upload` response, hard cut at last whitespace before the limit.

`truncatedRawDoc` is not available from the backend (raw text is not returned).
**Omit the detokenized version of the document context entirely** — the
`.detokenized` paragraph for this bubble is hidden permanently and the toggle
skips `.message.context` bubbles. This is acceptable: the document context
bubble exists to show *what the LLM sees*, which is always tokenized.

---

## Global Toggle Button

```html
<div id="chat-header">
  <h2>Conversation</h2>
  <button id="toggle-pii" data-mode="tokenized">🔒 Show real data</button>
</div>
```

- Default: `tokenized` → shows `{{TOKEN}}` placeholders, button reads `🔒 Show real data`
- Toggled: `detokenized` → shows real values, button reads `🔓 Hide real data`
- Applies to: `.message.user` question text, `.message.assistant` response blocks
- Does NOT apply to: `.message.system`, `.message.context` (always shows tokenized)

---

## Message Bubble Structure

### System prompt bubble
```html
<div class="message system">
  <span class="block-label">System Prompt</span>
  <p class="message-text">You are a helpful assistant…</p>
</div>
```

### Document context bubble
```html
<div class="message context">
  <span class="block-label">Document Context (sent to LLM)</span>
  <p class="tokenized">
    LAST WILL AND TESTAMENT OF {{PERSON_1}}, a resident of…
    <em class="truncation-notice">— truncated · full document sent to LLM —</em>
  </p>
</div>
```

### User bubble
```html
<div class="message user">
  <p class="tokenized">Who is the executor?</p>
  <p class="detokenized hidden">Who is the executor?</p>
</div>
```
Right-aligned, accent background. For most questions PII is unlikely, but the
dual-layer is included for consistency and for edge cases where the user types
a name into the question (tokenizer runs on the question too).

### Assistant response bubble
```html
<div class="message assistant">
  <span class="block-label">Response</span>
  <p class="tokenized">The executor is {{PERSON_2}}, residing at {{ADDRESS_1}}.</p>
  <p class="detokenized hidden">The executor is Jane Doe, residing at 123 Main St.</p>
</div>
```

---

## JS: Session Initialization (on upload success)

```js
function seedChatThread(uploadData) {
  // uploadData = { session_id, facts, system_prompt, tokenized_doc }
  chatThread.innerHTML = ''; // clear any previous session

  // 1. System prompt bubble
  const sysBubble = document.createElement('div');
  sysBubble.className = 'message system';
  sysBubble.innerHTML = `
    <span class="block-label">System Prompt</span>
    <p class="message-text">${escapeHtml(uploadData.system_prompt)}</p>
  `;
  chatThread.appendChild(sysBubble);

  // 2. Document context bubble (truncated)
  const TRUNCATE_AT = 400;
  const raw = uploadData.tokenized_doc;
  const truncated = raw.length > TRUNCATE_AT
    ? raw.slice(0, raw.lastIndexOf(' ', TRUNCATE_AT))
    : raw;
  const isTruncated = raw.length > TRUNCATE_AT;

  const ctxBubble = document.createElement('div');
  ctxBubble.className = 'message context';
  ctxBubble.innerHTML = `
    <span class="block-label">Document Context (sent to LLM)</span>
    <p class="tokenized">${highlightTokens(truncated)}${isTruncated
      ? '<br><em class="truncation-notice">— truncated · full document sent to LLM —</em>'
      : ''}</p>
  `;
  chatThread.appendChild(ctxBubble);
}
```

Call `seedChatThread(data)` inside `uploadFile()` on success, before
`setState('READY')`.

---

## JS: Rendering Q&A Messages

```js
function appendUserMessage(question, tokenizedQuestion) {
  const mode = toggleBtn.dataset.mode;
  const div = document.createElement('div');
  div.className = 'message user';
  div.innerHTML = `
    <p class="tokenized ${mode === 'detokenized' ? 'hidden' : ''}">${highlightTokens(tokenizedQuestion)}</p>
    <p class="detokenized ${mode === 'tokenized' ? 'hidden' : ''}">${escapeHtml(question)}</p>
  `;
  chatThread.appendChild(div);
  chatThread.scrollTop = chatThread.scrollHeight;
}

function appendAssistantMessage(data) {
  // data = { answer, tokenized_prompt, tokenized_answer }
  const mode = toggleBtn.dataset.mode;
  const div = document.createElement('div');
  div.className = 'message assistant';
  div.innerHTML = `
    <span class="block-label">Response</span>
    <p class="tokenized ${mode === 'detokenized' ? 'hidden' : ''}">${highlightTokens(data.tokenized_answer)}</p>
    <p class="detokenized ${mode === 'tokenized' ? 'hidden' : ''}">${escapeHtml(data.answer)}</p>
  `;
  chatThread.appendChild(div);
  chatThread.scrollTop = chatThread.scrollHeight;
}
```

`tokenizedQuestion` comes from `data.tokenized_prompt` returned by `/ask`.
Call order in `askQuestion()`:
1. `appendUserMessage(originalQuestion, data.tokenized_prompt)`
2. `appendAssistantMessage(data)`

---

## JS: Global Toggle Handler

```js
toggleBtn.addEventListener('click', () => {
  const next = toggleBtn.dataset.mode === 'tokenized' ? 'detokenized' : 'tokenized';
  toggleBtn.dataset.mode = next;
  toggleBtn.textContent = next === 'tokenized' ? '🔒 Show real data' : '🔓 Hide real data';

  // Only toggle user and assistant bubbles — skip system and context
  document.querySelectorAll('.message.user .tokenized, .message.assistant .tokenized')
    .forEach(el => el.classList.toggle('hidden', next === 'detokenized'));
  document.querySelectorAll('.message.user .detokenized, .message.assistant .detokenized')
    .forEach(el => el.classList.toggle('hidden', next === 'tokenized'));
});
```

---

## JS: Token Highlighting

```js
function escapeHtml(str) {
  return str.replace(/&/g,'&amp;').replace(/</g,'&lt;')
            .replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

function highlightTokens(text) {
  return escapeHtml(text).replace(
    /\{\{[A-Z_]+_\d+\}\}/g,
    m => `<mark class="pii-token">${m}</mark>`
  );
}
```

Use `highlightTokens()` for all `.tokenized` content.
Use `escapeHtml()` for all `.detokenized` content.

---

## CSS Additions

```css
#chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

#chat-thread {
  height: 460px;
  overflow-y: auto;
  border: 1px solid #ccc;
  border-radius: 6px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #fafafa;
}

/* System prompt */
.message.system {
  align-self: flex-start;
  background: #f0f0f0;
  border: 1px dashed #bbb;
  border-radius: 6px;
  padding: 10px 14px;
  max-width: 90%;
  font-size: 0.88em;
  color: #555;
}

/* Document context */
.message.context {
  align-self: flex-start;
  background: #f7f7f0;
  border: 1px solid #ddd;
  border-left: 3px solid #aaa;
  border-radius: 6px;
  padding: 10px 14px;
  max-width: 90%;
  font-size: 0.85em;
  font-family: monospace;
  white-space: pre-wrap;
  word-break: break-word;
}

.truncation-notice {
  display: block;
  margin-top: 6px;
  font-style: italic;
  color: #999;
  font-size: 0.9em;
  font-family: sans-serif;
}

/* User bubble */
.message.user {
  align-self: flex-end;
  background: #1a5276;
  color: white;
  border-radius: 12px 12px 2px 12px;
  padding: 10px 14px;
  max-width: 70%;
}

/* Assistant response */
.message.assistant {
  align-self: flex-start;
  background: white;
  border: 1px solid #ddd;
  border-left: 3px solid #1a5276;
  border-radius: 6px;
  padding: 10px 14px;
  max-width: 90%;
}

.block-label {
  font-size: 0.7em;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #888;
  display: block;
  margin-bottom: 4px;
}

mark.pii-token {
  background: #fef3cd;
  border: 1px solid #f0ad4e;
  border-radius: 3px;
  padding: 0 3px;
  font-family: monospace;
  font-size: 0.9em;
}

.hidden { display: none; }

#chat-input-row {
  display: flex;
  gap: 8px;
  margin-top: 10px;
}

#question-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #ccc;
  border-radius: 6px;
  font-size: 0.95em;
}

#ask-btn { white-space: nowrap; }

#toggle-pii {
  font-size: 0.85em;
  padding: 4px 10px;
  cursor: pointer;
}
```

---

## Updated Task List

### Phase F4 (replaces original F040–F041)

- [ ] **F040** — `handler/prompts.go`: extract `QASystemPrompt` constant shared
  by `upload.go` and `ask.go`. _(deps: T050)_

- [ ] **F041** — Backend: extend `UploadResponse` with `system_prompt` and
  `tokenized_doc`; populate from `QASystemPrompt` and `session.TokenText`.
  _(deps: F040, T061)_

- [ ] **F042** — Backend: extend `AskResponse` with `tokenized_prompt` and
  `tokenized_answer`; populate before detokenizing in `handler/ask.go`.
  _(deps: F040, T070)_

- [ ] **F043** — HTML: replace `#qa-section` with chat layout —
  `#chat-header` (h2 + toggle button), `#chat-thread` (scrollable div),
  `#chat-input-row` (input + ask button). _(deps: F010)_

- [ ] **F044** — CSS: add all styles from spec. _(deps: F043)_

- [ ] **F045** — JS: `escapeHtml`, `highlightTokens`, `seedChatThread`,
  `appendUserMessage`, `appendAssistantMessage`; wire `uploadFile()` to call
  `seedChatThread` on success; wire `askQuestion()` to call both append
  functions; Enter key on input. _(deps: F044, F041, F042)_

- [ ] **F046** — JS: global toggle handler — skip `.message.system` and
  `.message.context`, toggle only `.message.user` and `.message.assistant`
  `.tokenized` / `.detokenized` pairs. _(deps: F045)_

- [ ] **F047** — Verify end to end: load will sample → thread seeds with system
  prompt bubble + truncated document context bubble with `{{TOKEN}}` highlights
  → ask "Who is the executor?" → user bubble + assistant response appear →
  toggle shows real names across user + assistant bubbles only → context bubble
  unchanged. _(deps: F046, T071)_

---

## Dependency Summary (addendum only)

```
T061 ──┬── F040 ── F041
T070 ──┘         └── F042
F010 ── F043 ── F044 ── F045 (needs F041, F042) ── F046 ── F047
```