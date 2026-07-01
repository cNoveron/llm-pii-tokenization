// Package store holds the in-memory session state for the PoC. There is no
// TTL, eviction, or persistence — sessions live for the lifetime of the
// process.
package store

import "sync"

// Session is the state captured for one uploaded document.
type Session struct {
	ID        string            // UUID v4
	RawText   string            // Original extracted text (never sent to the LLM)
	TokenMap  map[string]string // token -> original value
	TokenText string            // PII-replaced text (safe to send to the LLM)
}

// Store is a concurrency-safe map of session id -> session.
type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// New returns an empty store ready for use.
func New() *Store {
	return &Store{sessions: make(map[string]*Session)}
}

// Set stores (or replaces) the session under id.
func (s *Store) Set(id string, session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = session
}

// Get returns the session for id and whether it was found.
func (s *Store) Get(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	return session, ok
}
