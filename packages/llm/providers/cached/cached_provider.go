package cached

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/stillmatic/gollum/packages/llm"
	"github.com/stillmatic/gollum/packages/llm/cache"
	"github.com/stillmatic/gollum/packages/llm/providers/cached/sqlitecache"

	"hash"

	_ "modernc.org/sqlite"
)

// CachedResponder implements the Responder interface with caching
type CachedResponder struct {
	underlying llm.Responder
	cache      cache.Cache
	hasher     hash.Hash
}

// NewLocalCachedResponder creates a new CachedResponder with a local SQLite cache
// For example, initialize an OpenAI provider and then wrap it with this cache.
func NewLocalCachedResponder(underlying llm.Responder, dbPath string) (*CachedResponder, error) {
	cache, err := sqlitecache.NewSQLiteCache(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	// we use sha256 to avoid pulling down the xxhash dep if you don't need to
	// these are small strings to cache so shouldn't make a big diff
	hasher := sha256.New()

	return &CachedResponder{
		underlying: underlying,
		cache:      cache,
		hasher:     hasher,
	}, nil
}

func (cr *CachedResponder) GenerateResponse(ctx context.Context, req llm.InferRequest) (string, error) {
	// Check cache
	cachedResponse, err := cr.cache.GetResponse(ctx, req)
	if err == nil {
		return cachedResponse, nil
	}

	// If not in cache, call underlying provider
	response, err := cr.underlying.GenerateResponse(ctx, req)
	if err != nil {
		return "", err
	}

	// Cache the result
	if err := cr.cache.SetResponse(ctx, req, response); err != nil {
		log.Printf("Failed to cache response: %v", err)
	}

	return response, nil
}

func (cr *CachedResponder) GenerateResponseAsync(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	// For async responses, we don't cache and just pass through to the underlying provider
	// TODO: think about if we should just cache final response and return immediately
	return cr.underlying.GenerateResponseAsync(ctx, req)
}

func (cr *CachedResponder) Close() error {
	return cr.cache.Close()
}

func (cr *CachedResponder) GetCacheStats() cache.CacheStats {
	return cr.cache.GetStats()
}
