package cached

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"log"

	"github.com/stillmatic/gollum/packages/llm"
	"github.com/stillmatic/gollum/packages/llm/cache"
	"github.com/stillmatic/gollum/packages/llm/providers/cached/sqlitecache"
)

// CachedEmbedder implements the llm.Embedder interface with caching
type CachedEmbedder struct {
	underlying llm.Embedder
	cache      cache.Cache
	hasher     hash.Hash
}

// NewLocalCachedEmbedder creates a new CachedEmbedder with a local SQLite cache
func NewLocalCachedEmbedder(underlying llm.Embedder, dbPath string) (*CachedEmbedder, error) {
	cache, err := sqlitecache.NewSQLiteCache(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return &CachedEmbedder{
		underlying: underlying,
		cache:      cache,
		hasher:     sha256.New(),
	}, nil
}

func (ce *CachedEmbedder) GenerateEmbedding(ctx context.Context, req llm.EmbedRequest) (*llm.EmbeddingResponse, error) {
	cachedEmbeddings := make([]llm.Embedding, 0, len(req.Input))
	uncachedIndices := make([]int, 0)
	uncachedInputs := make([]string, 0)

	// Check cache for each input string
	for i, input := range req.Input {
		embedding, err := ce.cache.GetEmbedding(ctx, req.ModelConfig.ModelName, input)
		if err == nil {
			cachedEmbeddings = append(cachedEmbeddings, llm.Embedding{Values: embedding})
		} else {
			uncachedIndices = append(uncachedIndices, i)
			uncachedInputs = append(uncachedInputs, input)
		}
	}

	// If all embeddings were cached, return immediately
	if len(uncachedInputs) == 0 {
		return &llm.EmbeddingResponse{
			Data: cachedEmbeddings,
		}, nil
	}

	// Generate embeddings for uncached inputs
	uncachedReq := llm.EmbedRequest{
		ModelConfig: req.ModelConfig,
		Input:       uncachedInputs,
	}
	uncachedResponse, err := ce.underlying.GenerateEmbedding(ctx, uncachedReq)
	if err != nil {
		return nil, err
	}

	// Cache the new embeddings
	for i, embedding := range uncachedResponse.Data {
		if err := ce.cache.SetEmbedding(ctx, req.ModelConfig.ModelName, uncachedInputs[i], embedding.Values); err != nil {
			log.Printf("Failed to cache embedding: %v", err)
		}
	}

	// Merge cached and new embeddings
	finalEmbeddings := make([]llm.Embedding, len(req.Input))
	cachedIndex, uncachedIndex := 0, 0
	for i := range req.Input {
		if contains(uncachedIndices, i) {
			finalEmbeddings[i] = uncachedResponse.Data[uncachedIndex]
			uncachedIndex++
		} else {
			finalEmbeddings[i] = cachedEmbeddings[cachedIndex]
			cachedIndex++
		}
	}

	return &llm.EmbeddingResponse{
		Data: finalEmbeddings,
	}, nil
}
func (ce *CachedEmbedder) Close() error {
	return ce.cache.Close()
}

func (ce *CachedEmbedder) GetCacheStats() cache.CacheStats {
	return ce.cache.GetStats()
}

func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
