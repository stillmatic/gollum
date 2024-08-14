package openai

import (
	"context"
	"github.com/stillmatic/gollum/packages/llm"
)

func NewDeepseekProvider(apiKey string) *DeepseekProvider {
	p := NewGenericProvider(apiKey, "https://api.deepseek.com/v1")

	return &DeepseekProvider{p, make(map[string]llm.InferMessage)}
}

// DeepseekProvider implements the CachableResponder and Responder interfaces.
// Because Deepseek caching is transparent to the end user, we don't need to do very much.
type DeepseekProvider struct {
	*Provider
	cache map[string]llm.InferMessage
}

// 	GenerateResponse(ctx context.Context, req InferRequest) (string, error)
//	GenerateResponseAsync(ctx context.Context, req InferRequest) (<-chan StreamDelta, error)
//CacheObject(ctx context.Context, key string, value InferMessage) error
//GetCachedObject(ctx context.Context, key string) (InferMessage, error)

func (p *DeepseekProvider) GenerateResponse(ctx context.Context, req llm.InferRequest) (string, error) {
	return p.Provider.GenerateResponse(ctx, req)
}

func (p *DeepseekProvider) GenerateResponseAsync(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	return p.Provider.GenerateResponseAsync(ctx, req)
}

func (p *DeepseekProvider) CacheObject(ctx context.Context, key string, value llm.InferMessage) error {
	p.cache[key] = value
	return nil
}

func (p *DeepseekProvider) GetCachedObject(ctx context.Context, key string) (llm.InferMessage, error) {
	return p.cache[key], nil
}
