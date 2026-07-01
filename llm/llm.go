package llm

import (
	"os"
)

// Complete sends a single-turn request with the given system prompt and user
// message to the configured LLM provider (Anthropic or Gemini) and returns
// the response. It defaults to Anthropic unless LLM_PROVIDER=gemini.
func Complete(systemPrompt, userMessage string) (string, error) {
	provider := os.Getenv("LLM_PROVIDER")

	if provider == "gemini" {
		return completeGemini(systemPrompt, userMessage)
	}

	// Default to Anthropic
	return completeAnthropic(systemPrompt, userMessage)
}
