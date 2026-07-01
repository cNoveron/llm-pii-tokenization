package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"trustate/llm"
	"trustate/pii"
	"trustate/store"
)

// AskRequest is the POST /ask request body.
type AskRequest struct {
	SessionID string `json:"session_id"`
	Question  string `json:"question"`
}

// AskResponse is the POST /ask response body (answer is de-tokenized).
type AskResponse struct {
	Answer          string `json:"answer"`
	TokenizedPrompt string `json:"tokenized_prompt"`
	TokenizedAnswer string `json:"tokenized_answer"`
}

// Ask handles POST /ask: look up the session, tokenize the question in the
// session's existing token space, ask the LLM against the tokenized document,
// and return the de-tokenized answer.
func Ask(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if req.SessionID == "" || req.Question == "" {
			writeError(w, http.StatusBadRequest, "session_id and question are required")
			return
		}

		session, ok := s.Get(req.SessionID)
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}

		// Tokenize the question, seeded from the document's token map so known
		// entities reuse their tokens and new PII gets fresh, non-colliding ones.
		tk := pii.NewFrom(session.TokenMap)
		questionText := tk.Apply(req.Question)
		combined := tk.Map()

		system := fmt.Sprintf("%s\n\nDocument:\n%s", QASystemPrompt, session.TokenText)
		reply, err := llm.Complete(system, questionText)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "LLM request failed")
			return
		}

		tokenizedAnswer := reply
		answer := pii.Detokenize(reply, combined)
		writeJSON(w, http.StatusOK, AskResponse{
			Answer:          answer,
			TokenizedPrompt: questionText,
			TokenizedAnswer: tokenizedAnswer,
		})
	}
}
