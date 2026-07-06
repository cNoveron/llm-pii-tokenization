// Command trustate is the PoC backend for the Trustate document intake
// assistant: it accepts a PDF upload, tokenizes PII, sends the sanitized text
// to Claude, and returns the de-tokenized response.
package main

import (
	"log"
	"net/http"
	"os"

	"trustate/handler"
	"trustate/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sessions := store.New()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /upload", handler.Upload(sessions))
	mux.HandleFunc("POST /ask", handler.Ask(sessions))
	mux.HandleFunc("POST /automate/discover", handler.Discover(sessions))
	mux.HandleFunc("POST /automate/retrieve/balances", handler.RetrieveBalances(sessions))
	mux.HandleFunc("POST /automate/retrieve/deeds", handler.RetrieveDeeds(sessions))
	mux.HandleFunc("POST /automate/retrieve/records", handler.RetrieveRecords(sessions))
	mux.HandleFunc("POST /automate/actions/probate", handler.ActionsProbate(sessions))
	mux.HandleFunc("POST /automate/actions/transfer", handler.ActionsTransfer(sessions))
	mux.HandleFunc("POST /automate/actions/creditors", handler.ActionsCreditors(sessions))
	mux.HandleFunc("POST /automate/generate-document", handler.GenerateDocument(sessions))
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /sample/{name}", handler.Sample("sample"))
	mux.Handle("GET /", http.FileServer(http.Dir("./static/")))

	log.Printf("trustate backend listening on :%s", port)
	if err := http.ListenAndServe(":"+port, cors(mux)); err != nil {
		log.Fatal(err)
	}
}

// cors applies permissive CORS suitable for the PoC and short-circuits
// preflight OPTIONS requests.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
