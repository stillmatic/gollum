//go:generate mockgen -source llm.go -destination internal/mocks/llm.go
package llm

import (
	"context"
)

type ProviderType string
type ModelType string

// yuck sorry
const (
	ModelTypeLLM       ModelType = "llm"
	ModelTypeEmbedding ModelType = "embedding"
)

type ModelConfig struct {
	ProviderType ProviderType
	ModelName    string
	BaseURL      string

	ModelType                        ModelType
	CentiCentsPerMillionInputTokens  int
	CentiCentsPerMillionOutputTokens int
}

// MessageOptions are options that can be passed to the model for generating a response.
// NB chua: these are the only ones I use, I assume others are useful too...
type MessageOptions struct {
	MaxTokens   int
	Temperature float32
}

type InferMessage struct {
	Content string
	Role    string
	Image   []byte

	ShouldCache bool
}

type InferRequest struct {
	Messages []InferMessage

	// ModelConfig describes the model to use for generating a response.
	ModelConfig ModelConfig
	// MessageOptions are options that can be passed to the model for generating a response.
	MessageOptions MessageOptions
}

type StreamDelta struct {
	Text string
	EOF  bool
}

type Responder interface {
	GenerateResponse(ctx context.Context, req InferRequest) (string, error)
	GenerateResponseAsync(ctx context.Context, req InferRequest) (<-chan StreamDelta, error)
}

type EmbedRequest struct {
	Input []string
	Image []byte

	// Prompt is an instruction applied to all the input strings in this request.
	// Ignored unless the model specifically supports it
	Prompt string

	ModelConfig ModelConfig
	// only supported for openai (matryoshka) models
	Dimensions int
}

type Embedding struct {
	Values []float32
}

type EmbeddingResponse struct {
	Data []Embedding
}

type Embedder interface {
	GenerateEmbedding(ctx context.Context, req EmbedRequest) (*EmbeddingResponse, error)
}
