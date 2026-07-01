# Trustate Document Intake Assistant — Backend Implementation Plan

Derived from [`spec.md`](./spec.md). This plan translates the specification
into an ordered, file-by-file implementation strategy with the concrete
decisions needed to build it.

---

## 1. Architecture Overview

A single-process Go HTTP service, stdlib-only for transport, with three
in-process collaborators and one external call:

```
                    ┌─────────────────────────────────────────┐
   PDF upload  ───▶ │  main.go  (net/http mux + CORS)          │
   JSON ask    ───▶ │    ├─ handler/upload.go                  │
                    │    ├─ handler/ask.go                      │
                    │    ├─ handler/sample.go                   │
                    │    └─ handler/health.go                   │
                    └──────────┬───────────────┬───────────────┘
                               │               │
                 ┌─────────────▼──┐   ┌─────────▼──────────┐
                 │ pii.Tokenizer  │   │ store.Store        │
                 │ (regex detect) │   │ (in-mem, RWMutex)  │
                 └─────────────┬──┘   └────────────────────┘
                               │
                    ┌──────────▼───────────┐        ┌──────────────────┐
                    │ llm.Complete()       │ ─HTTP─▶ │ Anthropic API    │
                    │ (raw net/http POST)  │        │ /v1/messages     │
                    └──────────────────────┘        └──────────────────┘
```

**Request lifecycle (upload):** parse multipart → extract PDF text →
tokenize PII → persist session → LLM fact extraction on tokenized text →
de-tokenize facts → JSON response.

**Request lifecycle (ask):** look up session → tokenize question (seeded
from the session's existing token map) → build Q&A prompt with tokenized
document → LLM → de-tokenize answer → JSON response.

**Key invariant:** raw PII never leaves the process. Only tokenized text is
sent to the Anthropic API; the token→value map lives only in the in-memory
session store, and de-tokenization happens on the response path.

---

## 2. Tech Stack Decisions

| Concern | Decision | Rationale |
|---|---|---|
| HTTP routing | `net/http` `ServeMux` with Go 1.22 method+wildcard patterns (`"POST /upload"`, `"GET /sample/{name}"`) | Stdlib-only per spec; 1.22 routing removes the need for a framework or manual method checks. `go.mod` targets 1.26. |
| PDF text extraction | `github.com/ledongthuc/pdf` via `NewReader` + `GetPlainText` | Named in spec; reads from an in-memory `io.ReaderAt`, no temp files. |
| Sample PDF generation | `github.com/signintech/gopdf` in a one-shot generator | Already in `go.mod`; produces well-formed PDFs that `ledongthuc/pdf` can parse reliably. |
| LLM transport | Raw `net/http` POST to `https://api.anthropic.com/v1/messages` | Spec mandates raw HTTP (no SDK). Verified wire contract: headers `x-api-key`, `anthropic-version: 2023-06-01`, `Content-Type: application/json`. |
| Model params | `claude-sonnet-4-6`, `max_tokens: 1024`, `temperature: 0` | Per spec. Sonnet 4.6 still accepts `temperature`; `0` for deterministic structured extraction. Do **not** apply `omitempty` to `temperature` (would drop the `0`). |
| Session IDs | UUID v4 via `crypto/rand` | Avoids adding `google/uuid`; 16 random bytes with version/variant bits set. |
| CORS | Inline middleware wrapping the mux | Spec: allow `*`, methods `GET, POST, OPTIONS`, header `Content-Type`. Middleware short-circuits `OPTIONS` with `204`. |
| Config | `os.Getenv` (`ANTHROPIC_API_KEY`, `PORT` default `8080`) | Per spec. |

---

## 3. Package & File Plan

Module `trustate` at repo root (per prior decision — the spec's
`trustate-demo/` tree is illustrative).

| File | Responsibility | Key symbols |
|---|---|---|
| `main.go` | Wire mux, CORS, routes, server start | `main`, `cors(next)` |
| `store/session.go` | In-memory session store | `Session`, `Store`, `New`, `Set`, `Get` |
| `pii/tokenizer.go` | Regex PII detect / tokenize / de-tokenize | `Tokenizer`, `New`, `NewFrom`, `Apply`, `Map`, `Tokenize`, `Detokenize` |
| `llm/claude.go` | Anthropic Messages API client | `Complete(system, user) (string, error)` |
| `handler/common.go` | Shared helpers | `writeJSON`, `writeError`, `extractPDFText`, `parseFacts`, `newSessionID` |
| `handler/upload.go` | `POST /upload` | `Upload(*store.Store) http.HandlerFunc`, `Fact`, `UploadResponse` |
| `handler/ask.go` | `POST /ask` | `Ask(*store.Store) http.HandlerFunc`, `AskRequest`, `AskResponse` |
| `handler/sample.go` | `GET /sample/{name}` | `Sample(dir string) http.HandlerFunc` |
| `handler/health.go` | `GET /health` | `Health` |
| `sample/gen/` | One-shot sample PDF generator (own `go.mod` or `cmd`) | `main`, `buildPDF` |
| `.env.example` | Documented env vars | — |
| `README.md` | Quick start, curl examples, PoC caveats | — |

---

## 4. Design Detail: PII Tokenizer

The spec's minimal interface (`Tokenize`/`Detokenize`) is kept, but a
`Tokenizer` value is added so `/ask` can extend the document's existing
token space without collisions.

- **Entities & order:** SSN → DOB → AMOUNT → PHONE → EMAIL → PERSON.
  Order matters: EMAIL before PERSON (so a name inside an address is
  consumed as one email token); SSN before PHONE (distinct digit shapes,
  but SSN wins first).
- **PERSON detection:** regex alternation built from a hardcoded
  `knownNames` list (mock quality per spec), sorted longest-first,
  `regexp.QuoteMeta`'d, wrapped in `\b(?:…)\b`. **The name list must stay
  in sync with the names embedded in the generated sample PDFs.**
- **Dedup & numbering:** a `TYPE|value → token` map ensures repeated values
  reuse one token; per-type counters produce `{{PERSON_1}}`,
  `{{PERSON_2}}`, etc.
- **`NewFrom(existing)`:** seeds counters (parsed from existing token names
  via `^\{\{([A-Z]+)_(\d+)\}\}$`) and the reverse map, so a question
  mentioning a document entity reuses its token and new PII continues the
  numbering — no `{{PERSON_1}}` collision between doc and question.
- **`Detokenize`:** replace tokens longest-first (defensive; the `}}`
  terminator already prevents prefix ambiguity).

---

## 5. Design Detail: LLM Client

Single function `Complete(systemPrompt, userMessage string) (string, error)`.

- Request body: `{model, max_tokens:1024, temperature:0, system, messages:[{role:"user",content:userMessage}]}`.
- Reads `ANTHROPIC_API_KEY` at call time; returns a clear error if unset.
- `http.Client` with a 60s timeout.
- Response: concatenate `.content[]` blocks where `type == "text"`.
- Non-200: parse the Anthropic error envelope (`error.message`) when
  present, else return status + raw body.

**Prompts** (verbatim from spec §LLM Client):
- Fact extraction system prompt → JSON array `[{label,value}]`.
- Q&A system prompt with `Document:\n{tokenized_document_text}` appended.

**Fact parsing:** the model may wrap JSON in prose/fences, so extract the
substring from the first `[` to the last `]` before `json.Unmarshal`; on
failure return an empty slice (PoC-tolerant, never 500 on malformed JSON).

---

## 6. Milestones (build & verify order)

1. **Skeleton** — `go.mod`, `main.go` with mux + CORS + `/health`. Verify:
   `go run .` then `curl /health` → `{"status":"ok"}`.
2. **Store** — `store/session.go`. Compile check.
3. **Tokenizer** — `pii/tokenizer.go` + a table-driven unit test on a
   fixed string (SSN, name, amount, email → expected tokens, round-trip
   de-tokenize). Verify offline (no network).
4. **Sample PDFs** — `sample/gen/` generator; run once to emit
   `sample/will.pdf`, `sample/poa.pdf`. Wire `GET /sample/{name}`.
5. **LLM client** — `llm/claude.go`. Verify with a throwaway `/health`-style
   probe or the first real `/upload`.
6. **Upload** — `handler/upload.go` end-to-end with a sample PDF.
7. **Ask** — `handler/ask.go` using the session id from step 6.
8. **Docs** — `.env.example`, `README.md` (quick start + 2 curl examples +
   PoC-vs-production paragraph).

---

## 7. Dependencies & Prerequisites

- Go 1.22+ (repo pins 1.26).
- `ANTHROPIC_API_KEY` in the environment for `/upload` and `/ask`.
- Modules (fetched manually): `github.com/ledongthuc/pdf` (runtime),
  `github.com/signintech/gopdf` + transitive `gofpdi`, `pkg/errors`
  (sample generation only). Run `go mod tidy` / `go mod download` before
  first build.

---

## 8. Testing & Verification Strategy

- **Unit (offline):** tokenizer round-trip and collision behavior; fact
  JSON extraction from noisy strings.
- **Integration (needs key):** `curl -F file=@sample/will.pdf /upload`
  returns a session id + non-empty facts; `curl /ask` with that id answers
  a question and the answer contains real (de-tokenized) names, never
  `{{…}}` tokens.
- **Manual PII check:** log or assert that `TokenText` sent to the LLM
  contains no raw SSN/name from the source document.

---

## 9. Out of Scope (inherited from spec)

Auth, persistence, real PII detection (Presidio/DLP), multi-turn history,
streaming, OCR, retry logic, TLS. No work planned for any of these.

---

## 10. Risks & Mitigations

| Risk | Mitigation |
|---|---|
| Generated sample PDF not parseable by `ledongthuc/pdf` | Use `gopdf` (well-formed output) rather than hand-rolled bytes; verify extraction in milestone 4. |
| `knownNames` drifts from sample content | Single source of truth: names live beside the generator; document the coupling in `pii/tokenizer.go`. |
| Model returns non-JSON facts | Bracket-substring extraction + tolerant empty-slice fallback. |
| Token collisions between document and question | `pii.NewFrom` seeds counters/reverse-map from the session's map. |
| `temperature:0` accidentally dropped | No `omitempty` on the field; covered in code review. |
