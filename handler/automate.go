package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"trustate/store"
)

// AutomateRequest represents the body of the automation stage requests.
type AutomateRequest struct {
	SessionID string `json:"session_id"`
}

// AutomateResponse represents the placeholder response for automation stages.
type AutomateResponse struct {
	Status string `json:"status"`
}

// Discover handles POST /automate/discover
func Discover(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AutomateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		log.Printf("[Discover] Received SessionID: %s", req.SessionID)

		// Look up the session in store
		if _, ok := s.Get(req.SessionID); !ok {
			log.Printf("[Discover] Warning: SessionID %s not found in store", req.SessionID)
		}

		writeJSON(w, http.StatusOK, AutomateResponse{Status: "not_implemented"})
	}
}

// Retrieve handles POST /automate/retrieve
func Retrieve(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AutomateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		log.Printf("[Retrieve] Received SessionID: %s", req.SessionID)

		// Look up the session in store
		if _, ok := s.Get(req.SessionID); !ok {
			log.Printf("[Retrieve] Warning: SessionID %s not found in store", req.SessionID)
		}

		writeJSON(w, http.StatusOK, AutomateResponse{Status: "not_implemented"})
	}
}

// Actions handles POST /automate/actions
func Actions(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AutomateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		log.Printf("[Actions] Received SessionID: %s", req.SessionID)

		// Look up the session in store
		if _, ok := s.Get(req.SessionID); !ok {
			log.Printf("[Actions] Warning: SessionID %s not found in store", req.SessionID)
		}

		writeJSON(w, http.StatusOK, AutomateResponse{Status: "not_implemented"})
	}
}
