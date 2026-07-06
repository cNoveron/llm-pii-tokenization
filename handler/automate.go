package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"trustate/llm"
	"trustate/pii"
	"trustate/store"
)

// AutomateRequest represents the body of the automation stage requests.
type AutomateRequest struct {
	SessionID string `json:"session_id"`
}

// AutomateResponse represents the response for automation stages.
type AutomateResponse struct {
	Status          string `json:"status"`
	Answer          string `json:"answer"`
	TokenizedAnswer string `json:"tokenized_answer"`
}

const AutomateSystemPrompt = `You are a helpful assistant for legal staff reviewing estate documents.
The document uses privacy tokens like {{PERSON_1}} instead of real names.
Preserve tokens exactly as-is in your responses.
Be concise and return your response as a clean, bulleted list. Do not output conversational filler.`

// Discover handles POST /automate/discover
func Discover(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AutomateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		log.Printf("[Discover] Received SessionID: %s", req.SessionID)

		session, ok := s.Get(req.SessionID)
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}

		system := fmt.Sprintf("%s\n\nDocument:\n%s", AutomateSystemPrompt, session.TokenText)
		prompt := "Scan the document for assets, bank accounts, or properties, and list them. Return a concise, bulleted list of all assets found. If no assets are found, state that."

		reply, err := llm.Complete(system, prompt)
		if err != nil {
			log.Printf("[Discover] LLM failed: %v", err)
			writeError(w, http.StatusInternalServerError, "LLM request failed")
			return
		}

		answer := pii.Detokenize(reply, session.TokenMap)
		writeJSON(w, http.StatusOK, AutomateResponse{
			Status:          "success",
			Answer:          answer,
			TokenizedAnswer: reply,
		})
	}
}

// RetrieveBalances handles POST /automate/retrieve/balances
func RetrieveBalances(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AutomateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		log.Printf("[RetrieveBalances] Received SessionID: %s", req.SessionID)

		session, ok := s.Get(req.SessionID)
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}

		system := fmt.Sprintf("%s\n\nDocument:\n%s", AutomateSystemPrompt, session.TokenText)
		prompt := "Scan the document for any bank, brokerage, or financial accounts. List each account name/institution mentioned, and state that we need to contact them to retrieve the date-of-death balance. Return a concise, bulleted list."

		reply, err := llm.Complete(system, prompt)
		if err != nil {
			log.Printf("[RetrieveBalances] LLM failed: %v", err)
			writeError(w, http.StatusInternalServerError, "LLM request failed")
			return
		}

		answer := pii.Detokenize(reply, session.TokenMap)
		writeJSON(w, http.StatusOK, AutomateResponse{
			Status:          "success",
			Answer:          answer,
			TokenizedAnswer: reply,
		})
	}
}

// RetrieveDeeds handles POST /automate/retrieve/deeds
func RetrieveDeeds(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AutomateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		log.Printf("[RetrieveDeeds] Received SessionID: %s", req.SessionID)

		session, ok := s.Get(req.SessionID)
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}

		system := fmt.Sprintf("%s\n\nDocument:\n%s", AutomateSystemPrompt, session.TokenText)
		prompt := "Scan the document for any real property or physical estate mentions. State the need to lookup deed and title records to verify ownership and check for outstanding liens. Return a concise, bulleted list."

		reply, err := llm.Complete(system, prompt)
		if err != nil {
			log.Printf("[RetrieveDeeds] LLM failed: %v", err)
			writeError(w, http.StatusInternalServerError, "LLM request failed")
			return
		}

		answer := pii.Detokenize(reply, session.TokenMap)
		writeJSON(w, http.StatusOK, AutomateResponse{
			Status:          "success",
			Answer:          answer,
			TokenizedAnswer: reply,
		})
	}
}

// RetrieveRecords handles POST /automate/retrieve/records
func RetrieveRecords(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AutomateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		log.Printf("[RetrieveRecords] Received SessionID: %s", req.SessionID)

		session, ok := s.Get(req.SessionID)
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}

		system := fmt.Sprintf("%s\n\nDocument:\n%s", AutomateSystemPrompt, session.TokenText)
		prompt := "Scan the document for personal details of the decedent and primary executor/heirs. State the requirement to request necessary vital records (such as death certificates, marriage licenses, or letters testamentary). Return a concise, bulleted list."

		reply, err := llm.Complete(system, prompt)
		if err != nil {
			log.Printf("[RetrieveRecords] LLM failed: %v", err)
			writeError(w, http.StatusInternalServerError, "LLM request failed")
			return
		}

		answer := pii.Detokenize(reply, session.TokenMap)
		writeJSON(w, http.StatusOK, AutomateResponse{
			Status:          "success",
			Answer:          answer,
			TokenizedAnswer: reply,
		})
	}
}

// ActionsProbate handles POST /automate/actions/probate
func ActionsProbate(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AutomateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		log.Printf("[ActionsProbate] Received SessionID: %s", req.SessionID)

		session, ok := s.Get(req.SessionID)
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}

		system := fmt.Sprintf("%s\n\nDocument:\n%s", AutomateSystemPrompt, session.TokenText)
		prompt := "Scan the document for the designated executor, witnesses, and heirs. Outline the specific information required to draft the Petition for Probate. Return a concise, bulleted list."

		reply, err := llm.Complete(system, prompt)
		if err != nil {
			log.Printf("[ActionsProbate] LLM failed: %v", err)
			writeError(w, http.StatusInternalServerError, "LLM request failed")
			return
		}

		answer := pii.Detokenize(reply, session.TokenMap)
		writeJSON(w, http.StatusOK, AutomateResponse{
			Status:          "success",
			Answer:          answer,
			TokenizedAnswer: reply,
		})
	}
}

// ActionsTransfer handles POST /automate/actions/transfer
func ActionsTransfer(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AutomateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		log.Printf("[ActionsTransfer] Received SessionID: %s", req.SessionID)

		session, ok := s.Get(req.SessionID)
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}

		system := fmt.Sprintf("%s\n\nDocument:\n%s", AutomateSystemPrompt, session.TokenText)
		prompt := "Identify the assets mentioned in the document and the designated beneficiaries. Outline the legal deeds, transfer notices, or beneficiary claims needed to transfer these assets. Return a concise, bulleted list."

		reply, err := llm.Complete(system, prompt)
		if err != nil {
			log.Printf("[ActionsTransfer] LLM failed: %v", err)
			writeError(w, http.StatusInternalServerError, "LLM request failed")
			return
		}

		answer := pii.Detokenize(reply, session.TokenMap)
		writeJSON(w, http.StatusOK, AutomateResponse{
			Status:          "success",
			Answer:          answer,
			TokenizedAnswer: reply,
		})
	}
}

// ActionsCreditors handles POST /automate/actions/creditors
func ActionsCreditors(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AutomateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		log.Printf("[ActionsCreditors] Received SessionID: %s", req.SessionID)

		session, ok := s.Get(req.SessionID)
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}

		system := fmt.Sprintf("%s\n\nDocument:\n%s", AutomateSystemPrompt, session.TokenText)
		prompt := "Identify any debt, mortgage, or creditor references in the document, and outline the notice and publishing requirements to notify creditors of the estate administration. Return a concise, bulleted list."

		reply, err := llm.Complete(system, prompt)
		if err != nil {
			log.Printf("[ActionsCreditors] LLM failed: %v", err)
			writeError(w, http.StatusInternalServerError, "LLM request failed")
			return
		}

		answer := pii.Detokenize(reply, session.TokenMap)
		writeJSON(w, http.StatusOK, AutomateResponse{
			Status:          "success",
			Answer:          answer,
			TokenizedAnswer: reply,
		})
	}
}

// GenerateDocumentRequest represents the body of the document generation request.
type GenerateDocumentRequest struct {
	SessionID string `json:"session_id"`
	StageID   string `json:"stage_id"`
}

// GenerateDocumentResponse represents the response for document generation.
type GenerateDocumentResponse struct {
	Status           string `json:"status"`
	DocumentTitle    string `json:"document_title"`
	Content          string `json:"content"`
	TokenizedContent string `json:"tokenized_content"`
}

// GenerateDocument handles POST /automate/generate-document
func GenerateDocument(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req GenerateDocumentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		log.Printf("[GenerateDocument] Received SessionID: %s, StageID: %s", req.SessionID, req.StageID)

		session, ok := s.Get(req.SessionID)
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}

		var docTitle, prompt string
		switch req.StageID {
		case "1":
			docTitle = "Asset Summary Report"
			prompt = "Based on the document details, write a professional Asset Summary Report. Outline all discovered assets, accounts, properties, and values, structured clearly with headings for Real Estate, Financial Accounts, and Personal Property. Format the document beautifully with clean paragraphs and line breaks."
		case "2a":
			docTitle = "Asset Verification Request Letter"
			prompt = "Based on the document details, write a formal Request Letter to a financial institution to verify date-of-death balances. It should be addressed from the Executor, requesting balance details, outstanding interest, and account status. Format the letter with placeholders for addresses and a formal signature block."
		case "2b":
			docTitle = "Title & Deed Investigation Request"
			prompt = "Based on the document details, draft a Title Search and Deed lookup request. It should list the properties identified and formally request the registrar of deeds to verify ownership, liens, and recent transfers. Use a clean, formal request letter template format."
		case "2c":
			docTitle = "Vital Records Request Application Letter"
			prompt = "Based on the document details, draft a request letter to the Department of Vital Statistics/Records to request certified copies of the death certificate. Include decedent details, date of event, and relationship of the applicant. Use a formal layout with necessary fields."
		case "3a":
			docTitle = "Draft Petition for Probate"
			prompt = "Based on the document details, draft a formal Petition for Probate. Include sections for petitioner details, decedent details, date and place of death, estimate of asset value, and list of heirs and beneficiaries. Format it as a pleading or formal court petition."
		case "3b":
			docTitle = "Asset Transfer Notice Draft"
			prompt = "Based on the document details, draft an Asset Transfer Notice or Instruction letter. It should direct the transfer of specified accounts or properties from the decedent's name to the designated beneficiaries or the estate's name. Use a clean, formal layout."
		case "3c":
			docTitle = "Draft Notice to Creditors"
			prompt = "Based on the document details, draft a formal Notice to Creditors for publication. State that administration has begun, provide the executor name and contact, and announce the deadline for creditors to present their claims. Keep it in standard legal publication notice format."
		default:
			writeError(w, http.StatusBadRequest, "invalid stage ID")
			return
		}

		system := fmt.Sprintf("%s\n\nDocument:\n%s", AutomateSystemPrompt, session.TokenText)

		reply, err := llm.Complete(system, prompt)
		if err != nil {
			log.Printf("[GenerateDocument] LLM failed: %v", err)
			writeError(w, http.StatusInternalServerError, "LLM request failed")
			return
		}

		answer := pii.Detokenize(reply, session.TokenMap)
		writeJSON(w, http.StatusOK, GenerateDocumentResponse{
			Status:           "success",
			DocumentTitle:    docTitle,
			Content:          answer,
			TokenizedContent: reply,
		})
	}
}

