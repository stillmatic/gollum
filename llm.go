//go:generate mockgen -source llm.go -destination internal/mocks/llm.go

package gollum

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type Completer interface {
	CreateCompletion(context.Context, openai.CompletionRequest) (openai.CompletionResponse, error)
}

type ChatCompleter interface {
	CreateChatCompletion(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

type Embedder interface {
	CreateEmbeddings(context.Context, openai.EmbeddingRequest) (openai.EmbeddingResponse, error)
}

type Moderator interface {
	Moderations(context.Context, openai.ModerationRequest) (openai.ModerationResponse, error)
}
