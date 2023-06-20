//go:generate mockgen -source llm.go -destination internal/mocks/llm.go

package gollum

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// LLM is an interface for a language model
// We force all models to implement the OpenAI interface (which openai.client already implements)
// We only add the methods we need to use
type LLM interface {
	CreateCompletion(context.Context, openai.CompletionRequest) (openai.CompletionResponse, error)
	CreateChatCompletion(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
	CreateEmbeddings(context.Context, openai.EmbeddingRequest) (openai.EmbeddingResponse, error)
}

// LLMWithRetry is a wrapper around an LLM that retries N times
// with exponential backoff.
type LLMWithRetry struct {
	llm LLM
	N   int
}

func NewLLMWithRetry(llm LLM, n int) *LLMWithRetry {
	return &LLMWithRetry{llm: llm, N: n}
}
