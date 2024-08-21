package cache

import (
	"context"

	"github.com/stillmatic/gollum/packages/llm"
)

// Cache defines the interface for caching LLM responses and embeddings
type Cache interface {
	// GetResponse retrieves a cached response for a given request
	GetResponse(ctx context.Context, req llm.InferRequest) (string, error)

	// SetResponse stores a response for a given request
	SetResponse(ctx context.Context, req llm.InferRequest, response string) error

	// GetEmbedding retrieves a cached embedding for a given input and model config
	GetEmbedding(ctx context.Context, modelConfig string, input string) ([]float32, error)

	// SetEmbedding stores an embedding for a given input and model config
	SetEmbedding(ctx context.Context, modelConfig string, input string, embedding []float32) error

	// Close closes the cache, releasing any resources
	Close() error

	// GetStats returns cache statistics (e.g., number of requests, cache hits)
	GetStats() CacheStats
}

// CacheStats represents cache usage statistics
type CacheStats struct {
	NumRequests  int
	NumCacheHits int
}
