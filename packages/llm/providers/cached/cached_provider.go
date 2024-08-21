package cached

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/stillmatic/gollum/packages/llm"
	"log"

	"hash"
	_ "modernc.org/sqlite"
)

// CachedResponder implements the Responder interface with caching
// TODO: add a Cache interface to allow for different cache implementations
type CachedResponder struct {
	underlying   llm.Responder
	db           *sql.DB
	hasher       hash.Hash
	numRequests  int
	numCacheHits int
}

// CachedEmbedder implements the llm.Embedder interface with caching
type CachedEmbedder struct {
	underlying   llm.Embedder
	db           *sql.DB
	hasher       hash.Hash
	numRequests  int
	numCacheHits int
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

	// we use sha256 to avoid pulling down the xxhash dep if you don't need to
	// these are small strings to cache so shouldn't make a big diff
	hasher := sha256.New()

	return &CachedResponder{
		underlying: underlying,
		db:         db,
		hasher:     hasher,
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
		hasher:     sha256.New(),
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
	// set to WAL mode for better performance
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	return err
}

func initEmbedderDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS embedding_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			model_config TEXT,
			input_string TEXT,
			embedding BLOB,
			UNIQUE(model_config, input_string)
		);
	`)
	if err != nil {
		return err
	}
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
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
	cr.numRequests++
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	hashedRequest := cr.hasher.Sum(requestJSON)

	var response string
	err = cr.db.QueryRow("SELECT response FROM response_cache WHERE request = ?", hashedRequest).Scan(&response)
	if err != nil {
		return "", err
	}
	cr.numCacheHits++

	return response, nil
}

func (cr *CachedResponder) cacheResponse(req llm.InferRequest, response string) error {
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}
	hashedRequest := cr.hasher.Sum(requestJSON)

	_, err = cr.db.Exec("INSERT INTO response_cache (request, response) VALUES (?, ?)", hashedRequest, response)
	return err
}

func (cr *CachedResponder) Close() error {
	return cr.db.Close()
}

func (cr *CachedResponder) GetCacheStats() (int, int) {
	return cr.numRequests, cr.numCacheHits
}

func (ce *CachedEmbedder) GenerateEmbedding(ctx context.Context, req llm.EmbedRequest) (*llm.EmbeddingResponse, error) {
	cachedEmbeddings := make([]llm.Embedding, 0, len(req.Input))
	uncachedIndices := make([]int, 0)
	uncachedInputs := make([]string, 0)

	// Check cache for each input string
	for i, input := range req.Input {
		// we use ModelName as a key, I think it's fine
		embedding, err := ce.getCachedEmbedding(req.ModelConfig.ModelName, input)
		if err == nil {
			cachedEmbeddings = append(cachedEmbeddings, llm.Embedding{embedding})
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
		if err := ce.cacheEmbedding(req.ModelConfig.ModelName, uncachedInputs[i], embedding.Values); err != nil {
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

func (ce *CachedEmbedder) getCachedEmbedding(modelConfig string, input string) ([]float32, error) {
	ce.numRequests++
	var embeddingBlob []byte
	err := ce.db.QueryRow("SELECT embedding FROM embedding_cache WHERE model_config = ? AND input_string = ?", modelConfig, input).Scan(&embeddingBlob)
	if err != nil {
		return nil, err
	}
	ce.numCacheHits++

	var embedding []float32
	err = json.Unmarshal(embeddingBlob, &embedding)
	if err != nil {
		return nil, err
	}

	return embedding, nil
}

func (ce *CachedEmbedder) cacheEmbedding(modelConfig string, input string, embedding []float32) error {
	embeddingBlob, err := json.Marshal(embedding)
	if err != nil {
		return err
	}

	_, err = ce.db.Exec("INSERT OR REPLACE INTO embedding_cache (model_config, input_string, embedding) VALUES (?, ?, ?)", modelConfig, input, embeddingBlob)
	return err
}

func (ce *CachedEmbedder) Close() error {
	return ce.db.Close()
}

func (ce *CachedEmbedder) GetCacheStats() (int, int) {
	return ce.numRequests, ce.numCacheHits
}

func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
