package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// allowedSamples restricts which sample files may be served.
var allowedSamples = map[string]bool{"will": true, "poa": true}

// Sample serves one of the hardcoded sample PDFs from dir. It lets the frontend
// offer "load sample doc" without a real file picker.
func Sample(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if !allowedSamples[name] {
			writeError(w, http.StatusNotFound, "unknown sample")
			return
		}
		data, err := os.ReadFile(filepath.Join(dir, name+".pdf"))
		if err != nil {
			writeError(w, http.StatusNotFound, "sample not found; run the sample generator first")
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, name))
		_, _ = w.Write(data)
	}
}
