package interfaces

import "context"

// Message is a single turn in a conversation.
type Message struct {
	Role    string `json:"role"`    // "user" | "assistant" | "system"
	Content string `json:"content"`
}

// ContextManager trims or summarizes conversation history to fit the model window.
type ContextManager interface {
	// Manage receives the full history and returns a version safe to send to the LLM.
	Manage(ctx context.Context, history []Message, modelWindowTokens int) ([]Message, error)
}
