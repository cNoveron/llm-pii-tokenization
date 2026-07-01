# Trustate Document Intake Assistant — Backend (PoC)

A small Go HTTP backend that accepts a PDF upload, tokenizes PII, sends the
sanitized text to Claude, and returns the de-tokenized response. See
[`spec/001-backend/spec.md`](spec/001-backend/spec.md) for the full
specification and [`plan.md`](spec/001-backend/plan.md) /
[`tasks.md`](spec/001-backend/tasks.md) for the implementation plan.

## Quick start

```bash
# 1. Resolve dependencies (only github.com/ledongthuc/pdf is needed at runtime).
go mod tidy

# 2. Generate the two sample PDFs (writes sample/will.pdf and sample/poa.pdf).
go run ./sample/gen

# 3. Set your API key and run.
export ANTHROPIC_API_KEY=sk-ant-...
go run .
# -> trustate backend listening on :8080
# The frontend UI is then available at http://localhost:8080
```

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | yes | Anthropic API key (used by `POST /upload` and `POST /ask`). |
| `PORT` | no | HTTP listen port. Defaults to `8080`. |

See [`.env.example`](.env.example).

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/upload` | Upload a PDF (`multipart/form-data`, field `file`, max 5 MB). Returns a `session_id` and extracted facts. |
| `POST` | `/ask` | Ask a question about a session's document (`application/json`). |
| `GET` | `/health` | Liveness check → `{"status":"ok"}`. |
| `GET` | `/sample/{name}` | Download a sample PDF (`will` or `poa`). |

## curl examples

Upload a document and capture the session id:

```bash
SESSION_ID=$(curl -s -F "file=@sample/will.pdf" http://localhost:8080/upload \
  | tee /dev/stderr | sed -n 's/.*"session_id":"\([^"]*\)".*/\1/p')
```

Ask a question about it:

```bash
curl -s -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d "{\"session_id\":\"$SESSION_ID\",\"question\":\"Who is the executor and what does each beneficiary receive?\"}"
```

## Running the tests

The PII tokenizer has offline unit tests (no network / API key needed):

```bash
go test ./pii
```

## PoC shortcuts vs. production

This is a proof of concept and takes deliberate shortcuts that a production
system would not. **PII detection is naive regex plus a hardcoded name list**
(`pii/tokenizer.go`); real de-identification would use a purpose-built detector
such as Microsoft Presidio or Google Cloud DLP with an NER model, and would
handle addresses, org names, and fuzzy matches the regex misses. **Session
state is entirely in-memory** (`store/session.go`) with no TTL, so it is lost on
restart and cannot scale beyond one process; production would use a database or
cache (e.g. Postgres or Redis) with expiry. **There is no authentication,
authorization, TLS, rate limiting, retry logic, or input validation beyond a
file-size cap** — all of which are prerequisites before this handles real client
documents. The privacy guarantee that raw PII never reaches the LLM does hold in
this PoC (only tokenized text is sent, and de-tokenization happens locally on the
response), but everything around that guarantee is demo-grade.

> Note: the sample-PDF generator (`sample/gen`) uses only the Go standard
> library and the built-in Helvetica font, so no PDF library or font file is
> required to produce the fixtures. `go mod tidy` will therefore prune the
> `gopdf`/`gofpdi` modules if they were previously listed — the only runtime
> dependency is `github.com/ledongthuc/pdf`.
