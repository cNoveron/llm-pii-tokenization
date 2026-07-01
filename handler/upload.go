package handler

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"trustate/llm"
	"trustate/pii"
	"trustate/store"
)

const maxUploadBytes = 5 << 20 // 5 MB

// Fact is a single extracted key/value pair (values are de-tokenized).
type Fact struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// UploadResponse is the POST /upload response body.
type UploadResponse struct {
	SessionID    string `json:"session_id"`
	Facts        []Fact `json:"facts"`
	SystemPrompt string `json:"system_prompt"`
	TokenizedDoc string `json:"tokenized_doc"`
}

const factSystemPrompt = `You are an assistant that extracts key facts from estate law documents.
The document uses privacy tokens like {{PERSON_1}} instead of real names.
Preserve tokens exactly as-is in your output.
Respond ONLY with a JSON array: [{"label":"...","value":"..."}]
Extract: Testator, Executor, Beneficiaries, Estate value, Document date, Document type.
Omit any field not found in the document.`

// Upload handles POST /upload: parse a PDF, tokenize PII, store the session,
// ask the LLM for structured facts, and return de-tokenized facts.
func Upload(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Cap the whole request body (form overhead + file).
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+(1<<20))
		if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
			writeError(w, http.StatusBadRequest, "invalid or oversized multipart form")
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, "missing file field")
			return
		}
		defer file.Close()

		buf := &bytes.Buffer{}
		n, err := io.CopyN(buf, file, maxUploadBytes+1)
		if err != nil && err != io.EOF {
			writeError(w, http.StatusBadRequest, "could not read file")
			return
		}
		if n > maxUploadBytes {
			writeError(w, http.StatusBadRequest, "file exceeds 5MB limit")
			return
		}

		text, err := extractPDFText(buf.Bytes())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not parse PDF")
			return
		}

		tokenText, tokenMap := pii.Tokenize(text)

		id := newSessionID()
		s.Set(id, &store.Session{
			ID:        id,
			RawText:   text,
			TokenMap:  tokenMap,
			TokenText: tokenText,
		})

		reply, err := llm.Complete(factSystemPrompt, tokenText)
		if err != nil {
			log.Printf("LLM request failed: %v", err)
			writeError(w, http.StatusInternalServerError, "LLM request failed")
			return
		}

		facts := parseFacts(reply)
		for i := range facts {
			facts[i].Label = pii.Detokenize(facts[i].Label, tokenMap)
			facts[i].Value = pii.Detokenize(facts[i].Value, tokenMap)
		}

		writeJSON(w, http.StatusOK, UploadResponse{
			SessionID:    id,
			Facts:        facts,
			SystemPrompt: QASystemPrompt,
			TokenizedDoc: tokenText,
		})
	}
}
