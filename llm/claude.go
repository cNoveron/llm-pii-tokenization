// Package llm is a minimal Anthropic Messages API client built on raw
// net/http (no SDK), per the spec. It exposes a single Complete function.
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	apiURL     = "https://api.anthropic.com/v1/messages"
	model      = "claude-sonnet-4-6"
	apiVersion = "2023-06-01"
	maxTokens  = 1024
)

var httpClient = &http.Client{Timeout: 60 * time.Second}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type request struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	// Temperature is intentionally NOT omitempty: 0 is a meaningful value
	// (deterministic output for structured extraction) and must be sent.
	Temperature float64   `json:"temperature"`
	System      string    `json:"system,omitempty"`
	Messages    []message `json:"messages"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type response struct {
	Content    []contentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
}

type apiError struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// completeAnthropic sends a single-turn request with the given system prompt and user
// message and returns the concatenated text of the response using Anthropic's API.
func completeAnthropic(systemPrompt, userMessage string) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}

	body, err := json.Marshal(request{
		Model:       model,
		MaxTokens:   maxTokens,
		Temperature: 0,
		System:      systemPrompt,
		Messages:    []message{{Role: "user", Content: userMessage}},
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", apiVersion)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call anthropic: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var e apiError
		if json.Unmarshal(raw, &e) == nil && e.Error.Message != "" {
			return "", fmt.Errorf("anthropic API %d: %s", resp.StatusCode, e.Error.Message)
		}
		return "", fmt.Errorf("anthropic API %d: %s", resp.StatusCode, string(raw))
	}

	var out response
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	var sb strings.Builder
	for _, block := range out.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	return sb.String(), nil
}
