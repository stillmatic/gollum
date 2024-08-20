package cached

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/stillmatic/gollum/packages/llm"
	"log"

	_ "modernc.org/sqlite"
)

// CachedResponder implements the Responder interface with caching
type CachedResponder struct {
	underlying llm.Responder
	db         *sql.DB
}

// CachedEmbedder implements the llm.Embedder interface with caching
type CachedEmbedder struct {
	underlying llm.Embedder
	db         *sql.DB
}

// NewLocalCachedResponder creates a new CachedResponder with a local SQLite cache
// For example, initialize an OpenAI provider and then wrap it with this cache.
func NewLocalCachedResponder(underlying llm.Responder, dbPath string) (*CachedResponder, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := initResponderDB(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &CachedResponder{
		underlying: underlying,
		db:         db,
	}, nil
}

// NewLocalCachedEmbedder creates a new CachedEmbedder with a local SQLite cache
func NewLocalCachedEmbedder(underlying llm.Embedder, dbPath string) (*CachedEmbedder, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := initEmbedderDB(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &CachedEmbedder{
		underlying: underlying,
		db:         db,
	}, nil
}

func initResponderDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS response_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			request BLOB,
			response TEXT
		);
	`)
	return err
}

func initEmbedderDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS embedding_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			request BLOB,
			response BLOB
		);
	`)
	return err
}

func (cr *CachedResponder) GenerateResponse(ctx context.Context, req llm.InferRequest) (string, error) {
	// Check cache
	cachedResponse, err := cr.getCachedResponse(req)
	if err == nil {
		return cachedResponse, nil
	}

	// If not in cache, call underlying provider
	response, err := cr.underlying.GenerateResponse(ctx, req)
	if err != nil {
		return "", err
	}

	// Cache the result
	if err := cr.cacheResponse(req, response); err != nil {
		log.Printf("Failed to cache response: %v", err)
	}

	return response, nil
}

func (cr *CachedResponder) GenerateResponseAsync(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	// For async responses, we don't cache and just pass through to the underlying provider
	// TODO: think about if we should just cache final response and return immediately
	return cr.underlying.GenerateResponseAsync(ctx, req)
}

func (cr *CachedResponder) getCachedResponse(req llm.InferRequest) (string, error) {
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	var response string
	err = cr.db.QueryRow("SELECT response FROM response_cache WHERE request = ?", requestJSON).Scan(&response)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (cr *CachedResponder) cacheResponse(req llm.InferRequest, response string) error {
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = cr.db.Exec("INSERT INTO response_cache (request, response) VALUES (?, ?)", requestJSON, response)
	return err
}

func (cr *CachedResponder) Close() error {
	return cr.db.Close()
}

func (ce *CachedEmbedder) GenerateEmbedding(ctx context.Context, req llm.EmbedRequest) (*llm.EmbeddingResponse, error) {
	// Check cache
	cachedEmbedding, err := ce.getCachedEmbedding(req)
	if err == nil {
		return cachedEmbedding, nil
	}

	// If not in cache, call underlying embedder
	embedding, err := ce.underlying.GenerateEmbedding(ctx, req)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := ce.cacheEmbedding(req, embedding); err != nil {
		log.Printf("Failed to cache embedding: %v", err)
	}

	return embedding, nil
}

func (ce *CachedEmbedder) getCachedEmbedding(req llm.EmbedRequest) (*llm.EmbeddingResponse, error) {
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	var responseJSON []byte
	err = ce.db.QueryRow("SELECT response FROM embedding_cache WHERE request = ?", requestJSON).Scan(&responseJSON)
	if err != nil {
		return nil, err
	}

	var embedding llm.EmbeddingResponse
	err = json.Unmarshal(responseJSON, &embedding)
	if err != nil {
		return nil, err
	}

	return &embedding, nil
}

func (ce *CachedEmbedder) cacheEmbedding(req llm.EmbedRequest, embedding *llm.EmbeddingResponse) error {
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	responseJSON, err := json.Marshal(embedding)
	if err != nil {
		return err
	}

	_, err = ce.db.Exec("INSERT INTO embedding_cache (request, response) VALUES (?, ?)", requestJSON, responseJSON)
	return err
}

func (ce *CachedEmbedder) Close() error {
	return ce.db.Close()
}
