package handler

import "net/http"

// Health responds 200 with {"status":"ok"}. No logic.
func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
