# Trustate Document Intake Assistant — Backend Tasks

Ordered task breakdown derived from [`plan.md`](./plan.md). Tasks are grouped
by phase (following the plan's milestones). `[P]` = parallelizable with other
`[P]` tasks in the same phase. Each task lists its dependencies by ID.

Legend: `[ ]` pending · `[~]` in progress · `[x]` done

---

## Phase 0 — Project Setup

- [ ] **T001** — Confirm `go.mod` (module `trustate`, Go 1.22+) and fetch
  dependencies: `go get github.com/ledongthuc/pdf`, `go get
  github.com/signintech/gopdf`, then `go mod tidy`. _(deps: none)_
- [ ] **T002** — Create `.env.example` documenting `ANTHROPIC_API_KEY` and
  `PORT` (default `8080`). _(deps: none)_ `[P]`

## Phase 1 — HTTP Skeleton

- [ ] **T010** — `handler/common.go`: `writeJSON(w, status, v)` and
  `writeError(w, status, msg)` (JSON `{"error": msg}`). _(deps: T001)_
- [ ] **T011** — `handler/health.go`: `Health` handler → `200 {"status":"ok"}`.
  _(deps: T010)_ `[P]`
- [ ] **T012** — `main.go`: `ServeMux` with `GET /health`, `cors` middleware
  (allow `*`; methods `GET, POST, OPTIONS`; header `Content-Type`; `204` on
  `OPTIONS`), read `PORT`, `ListenAndServe`. _(deps: T011)_
- [ ] **T013** — Verify skeleton: `go run .` then
  `curl localhost:8080/health` returns `{"status":"ok"}`. _(deps: T012)_

## Phase 2 — Session Store

- [ ] **T020** — `store/session.go`: `Session{ID,RawText,TokenMap,TokenText}`,
  `Store{mu sync.RWMutex, sessions map}`, `New`, `Set`, `Get`. _(deps: T001)_
- [ ] **T021** — Instantiate a single `*store.Store` in `main.go`; thread it to
  handler constructors. _(deps: T020, T012)_

## Phase 3 — PII Tokenizer

- [ ] **T030** — `pii/tokenizer.go`: entity regexes (SSN, DOB, AMOUNT, PHONE,
  EMAIL) and `knownNames` list; build `personRe` alternation (longest-first,
  `QuoteMeta`, `\b(?:…)\b`) in `init`; define ordered `entities` slice.
  _(deps: T001)_
- [ ] **T031** — `Tokenizer` type + `New`, `NewFrom(existing)` (seed counters
  via `^\{\{([A-Z]+)_(\d+)\}\}$`, reverse map), `Apply(text)`, `Map()`.
  _(deps: T030)_
- [ ] **T032** — Package funcs `Tokenize(text)` (wraps `New().Apply`) and
  `Detokenize(text, map)` (replace longest-first). _(deps: T031)_
- [ ] **T033** — `pii/tokenizer_test.go`: table-driven round-trip test (SSN +
  name + amount + email → expected tokens; de-tokenize restores original) and
  a `NewFrom` no-collision test. Run `go test ./pii` (offline). _(deps: T032)_

## Phase 4 — Sample PDFs

- [ ] **T040** — `sample/gen/`: generator (`buildPDF(lines)` via `gopdf`)
  emitting `sample/will.pdf` and `sample/poa.pdf`; embed a person name, SSN,
  DOB, dollar amounts, phone, email, and 2–3 paragraphs of prose. Names MUST
  match `pii.knownNames`. _(deps: T001, T030)_
- [ ] **T041** — Run the generator once to produce both PDFs. _(deps: T040)_
- [ ] **T042** — `handler/sample.go`: `Sample(dir)` → serve `will`/`poa` as
  `application/octet-stream`, `404` for unknown names. _(deps: T010)_
- [ ] **T043** — Register `GET /sample/{name}` in `main.go`; verify
  `curl -o out.pdf localhost:8080/sample/will` downloads a valid PDF.
  _(deps: T042, T041, T012)_

## Phase 5 — LLM Client

- [ ] **T050** — `llm/claude.go`: `Complete(system, user)` — build request
  (`claude-sonnet-4-6`, `max_tokens:1024`, `temperature:0` **no omitempty**),
  headers (`x-api-key`, `anthropic-version:2023-06-01`, `Content-Type`), 60s
  client, concatenate `content[].text`, parse Anthropic error envelope on
  non-200. _(deps: T001)_

## Phase 6 — Upload Endpoint

- [ ] **T060** — `handler/common.go`: `extractPDFText(data)` via
  `pdf.NewReader` + `GetPlainText`; `newSessionID()` (UUID v4 from
  `crypto/rand`); `parseFacts(s)` (first-`[`-to-last-`]` + tolerant fallback).
  _(deps: T010)_
- [ ] **T061** — `handler/upload.go`: `Fact`, `UploadResponse`, fact system
  prompt; `Upload(store)` — parse multipart (5MB cap), extract text, tokenize,
  store session, `llm.Complete`, `parseFacts`, de-tokenize values, respond.
  Error codes: `400` missing/oversized/bad file, `500` parse/LLM failure.
  _(deps: T060, T020, T032, T050)_
- [ ] **T062** — Register `POST /upload`; verify
  `curl -F file=@sample/will.pdf localhost:8080/upload` returns a session id +
  non-empty facts (needs `ANTHROPIC_API_KEY`). _(deps: T061, T021, T043)_

## Phase 7 — Ask Endpoint

- [ ] **T070** — `handler/ask.go`: `AskRequest`, `AskResponse`, Q&A system
  prompt (`Document:\n{TokenText}`); `Ask(store)` — decode JSON, `404` unknown
  session, `pii.NewFrom(session.TokenMap)` → `Apply(question)`, `llm.Complete`,
  `Detokenize(answer, combinedMap)`, respond. Error codes: `404` no session,
  `400` missing fields, `500` LLM failure. _(deps: T060, T020, T032, T050)_
- [ ] **T071** — Register `POST /ask`; verify a question against the session
  from T062 returns a de-tokenized answer (no `{{…}}` leakage). _(deps: T070,
  T021, T062)_

## Phase 8 — Documentation

- [ ] **T080** — `README.md`: `go run .` quick start, required env vars, two
  `curl` examples (upload + ask), and one paragraph on PoC shortcuts vs.
  production (regex PII, in-memory store, no auth). _(deps: T062, T071)_

---

## Dependency Summary

```
T001 ─┬─ T010 ─ T011 ─ T012 ─ T013
      │                  └─ T021
      ├─ T020 ─────────────┘
      ├─ T030 ─ T031 ─ T032 ─ T033
      │                └─ T040 ─ T041 ─ T043
      ├─ T042 ───────────────────┘
      └─ T050
T060 (needs T010) ─┬─ T061 ─ T062
                   └─ T070 ─ T071
T080 (needs T062, T071)
```

## Coverage Check

| Spec area | Task(s) |
|---|---|
| `POST /upload` flow + errors | T060–T062 |
| `POST /ask` flow + errors | T060, T070–T071 |
| `GET /health` | T011–T013 |
| `GET /sample/{name}` | T042–T043 |
| PII tokenizer (detect/tokenize/de-tokenize) | T030–T033 |
| LLM client (model, params, headers, prompts) | T050 |
| In-memory session store | T020–T021 |
| CORS middleware | T012 |
| Sample documents | T040–T041 |
| Env vars / config | T002, T050 |
| README requirements | T080 |
