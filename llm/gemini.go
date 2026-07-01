package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	geminiModel = "gemini-3.1-flash-lite"
)

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiSystemInstruction struct {
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	Temperature float64 `json:"temperature"`
}

type geminiRequest struct {
	SystemInstruction *geminiSystemInstruction `json:"system_instruction,omitempty"`
	Contents          []geminiContent          `json:"contents"`
	GenerationConfig  geminiGenerationConfig   `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

func completeGemini(systemPrompt, userMessage string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not set")
	}

	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", geminiModel, apiKey)

	reqBody := geminiRequest{
		SystemInstruction: &geminiSystemInstruction{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: userMessage}},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature: 0,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call gemini: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var e geminiResponse
		if json.Unmarshal(raw, &e) == nil && e.Error != nil && e.Error.Message != "" {
			return "", fmt.Errorf("gemini API %d: %s", resp.StatusCode, e.Error.Message)
		}
		return "", fmt.Errorf("gemini API %d: %s", resp.StatusCode, string(raw))
	}

	var out geminiResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(out.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in gemini response")
	}

	var sb strings.Builder
	for _, part := range out.Candidates[0].Content.Parts {
		sb.WriteString(part.Text)
	}

	return sb.String(), nil
}
