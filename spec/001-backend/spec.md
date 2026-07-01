# Trustate Document Intake Assistant — Backend Spec

## Overview

A Go HTTP backend that accepts a PDF upload, tokenizes PII, sends the
sanitized text to an LLM, and returns the de-tokenized response.
PoC scope: happy paths only, in-memory state, no auth, no persistence.

---

## Stack

- Language: Go 1.22+
- HTTP: `net/http` (stdlib only, no framework)
- PDF extraction: `github.com/ledongthuc/pdf`
- LLM: Anthropic Claude API (`claude-sonnet-4-6`) via raw `net/http` POST
- CORS: inline middleware (no library)
- Config: environment variables via `os.Getenv`

---

## Project Layout

```
trustate-demo/
├── main.go
├── handler/
│   ├── upload.go       # POST /upload
│   └── ask.go          # POST /ask
├── pii/
│   └── tokenizer.go    # Detect, tokenize, de-tokenize
├── llm/
│   └── claude.go       # Anthropic API client
├── store/
│   └── session.go      # In-memory session store
├── sample/
│   ├── will.pdf        # Sample estate document
│   └── poa.pdf         # Sample power of attorney
├── .env.example
└── README.md
```

---

## Environment Variables

| Variable | Description |
|---|---|
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `PORT` | HTTP listen port (default `8080`) |

---

## Data Models

### Session
```go
type Session struct {
    ID        string            // UUID v4
    RawText   string            // Original extracted text
    TokenMap  map[string]string // token → original value
    TokenText string            // PII-replaced text
}
```

### UploadResponse
```go
type UploadResponse struct {
    SessionID string   `json:"session_id"`
    Facts     []Fact   `json:"facts"`
}

type Fact struct {
    Label string `json:"label"` // e.g. "Executor"
    Value string `json:"value"` // de-tokenized value
}
```

### AskRequest / AskResponse
```go
type AskRequest struct {
    SessionID string `json:"session_id"`
    Question  string `json:"question"`
}

type AskResponse struct {
    Answer string `json:"answer"` // de-tokenized
}
```

---

## Endpoints

### POST /upload

**Content-Type:** `multipart/form-data`
**Field:** `file` (PDF, max 5MB)

**Flow:**
1. Parse multipart form, read `file` field into memory buffer
2. Extract plain text via `ledongthuc/pdf`
3. Run tokenizer → produce `TokenMap` + `TokenText`
4. Generate `session_id` (UUID)
5. Store session in memory map (keyed by session_id)
6. Call LLM: extract structured facts from `TokenText`
   - Prompt asks for JSON array of `{label, value}` pairs
   - Values in response will be tokens; de-tokenize before returning
7. Return `UploadResponse`

**Error responses:**
- `400` — missing file, unsupported type
- `500` — PDF parse failure, LLM failure

---

### POST /ask

**Content-Type:** `application/json`
**Body:** `AskRequest`

**Flow:**
1. Look up session by `session_id` — 404 if not found
2. Tokenize the user's question (it may contain PII)
3. Build prompt: system context (tokenized document) + user question (tokenized)
4. Call LLM
5. De-tokenize LLM response
6. Return `AskResponse`

**Error responses:**
- `404` — session not found
- `400` — missing fields
- `500` — LLM failure

---

### GET /health

Returns `200 OK` with `{"status":"ok"}`. No logic.

---

### GET /sample/{name}

Returns one of the hardcoded sample PDFs as `application/octet-stream`.
`name` is `will` or `poa`. Returns `404` for unknown names.
Allows the frontend to offer "load sample doc" without a real file picker.

---

## PII Tokenizer (`pii/tokenizer.go`)

### Entities to detect (regex-based, mock quality is fine)

| Entity | Token format | Example regex hint |
|---|---|---|
| SSN | `{{SSN_1}}` | `\d{3}-\d{2}-\d{4}` |
| Date of birth | `{{DOB_1}}` | `\b(0?[1-9]\|1[0-2])\/\d{1,2}\/\d{4}\b` |
| Dollar amount | `{{AMOUNT_1}}` | `\$[\d,]+(\.\d{2})?` |
| Phone number | `{{PHONE_1}}` | `\(?\d{3}\)?[-.\s]\d{3}[-.\s]\d{4}` |
| Email | `{{EMAIL_1}}` | standard email regex |
| Full name | `{{PERSON_1}}` | match against a small hardcoded list of names in sample docs |

Counter per entity type resets per session. Tokens are globally unique
within a session (e.g. two different people → `{{PERSON_1}}`, `{{PERSON_2}}`).

### Interface

```go
func Tokenize(text string) (tokenized string, tokenMap map[string]string)
func Detokenize(text string, tokenMap map[string]string) string
```

`tokenMap` maps token → original value (e.g. `"{{PERSON_1}}" → "John Smith"`).

---

## LLM Client (`llm/claude.go`)

Single function:

```go
func Complete(systemPrompt, userMessage string) (string, error)
```

- Model: `claude-sonnet-4-6`
- Max tokens: `1024`
- Temperature: `0` (deterministic, important for structured extraction)
- Auth header: `x-api-key: $ANTHROPIC_API_KEY`
- `anthropic-version: 2023-06-01`

### Fact extraction prompt (system)

```
You are an assistant that extracts key facts from estate law documents.
The document uses privacy tokens like {{PERSON_1}} instead of real names.
Preserve tokens exactly as-is in your output.
Respond ONLY with a JSON array: [{"label":"...","value":"..."}]
Extract: Testator, Executor, Beneficiaries, Estate value, Document date, Document type.
Omit any field not found in the document.
```

### Q&A prompt (system)

```
You are a helpful assistant for legal staff reviewing estate documents.
The document uses privacy tokens like {{PERSON_1}} instead of real names.
Preserve tokens exactly as-is in your responses.
Answer based only on the document provided. Be concise.

Document:
{tokenized_document_text}
```

---

## In-Memory Session Store (`store/session.go`)

```go
type Store struct {
    mu       sync.RWMutex
    sessions map[string]*Session
}

func (s *Store) Set(id string, session *Session)
func (s *Store) Get(id string) (*Session, bool)
```

Single global store instance. No TTL needed for PoC.

---

## CORS Middleware

Allow all origins (`*`), methods `GET, POST, OPTIONS`,
headers `Content-Type`. Applied to all routes in `main.go`.

---

## Sample Documents

Two hardcoded fake PDFs in `sample/`. Content should include:
- At least one person's full name
- An SSN
- Dollar amounts
- A date
- Enough prose to support 2–3 meaningful Q&A exchanges

Generate these once with any PDF writer or a short Go script using `gopdf`.

---

## What is explicitly out of scope

- Authentication
- Persistent storage (DB, Redis)
- Real PII detection (Presidio, GCP DLP)
- Multi-turn conversation history
- Streaming responses
- PDF OCR (scanned images)
- Error retry logic
- Input sanitization beyond file size cap
- HTTPS / TLS

---

## README must include

1. `go run .` quick start
2. Required env vars
3. Two `curl` examples (upload + ask)
4. One paragraph explicitly calling out PoC shortcuts vs. production approach
   (regex PII detection, in-memory store, no auth)