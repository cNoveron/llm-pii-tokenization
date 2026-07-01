// Package handler contains the HTTP handlers and their shared helpers.
package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ledongthuc/pdf"
)

// writeJSON writes v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error body: {"error": msg}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// extractPDFText reads plain text from an in-memory PDF.
func extractPDFText(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	tr, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if _, err := io.Copy(&sb, tr); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// newSessionID returns a random UUID v4 string.
func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// parseFacts extracts a JSON array of facts from the model's reply. The model
// is asked to return only a JSON array, but may wrap it in prose or fences, so
// we slice from the first '[' to the last ']'. On any failure it returns an
// empty slice rather than erroring (PoC-tolerant).
func parseFacts(s string) []Fact {
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start == -1 || end == -1 || end < start {
		return []Fact{}
	}
	var facts []Fact
	if err := json.Unmarshal([]byte(s[start:end+1]), &facts); err != nil {
		return []Fact{}
	}
	return facts
}
