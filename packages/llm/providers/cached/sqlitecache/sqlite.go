package sqlitecache

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash"

	"github.com/stillmatic/gollum/packages/llm"
	"github.com/stillmatic/gollum/packages/llm/cache"
	_ "modernc.org/sqlite"
)

type SQLiteCache struct {
	db           *sql.DB
	hasher       hash.Hash
	numRequests  int
	numCacheHits int
}

func NewSQLiteCache(dbPath string) (*SQLiteCache, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := initDB(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &SQLiteCache{
		db:     db,
		hasher: sha256.New(),
	}, nil
}

func (c *SQLiteCache) GetResponse(ctx context.Context, req llm.InferRequest) (string, error) {
	c.numRequests++
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	hashedRequest := c.hasher.Sum(requestJSON)

	var response string
	err = c.db.QueryRowContext(ctx, "SELECT response FROM response_cache WHERE request = ?", hashedRequest).Scan(&response)
	if err != nil {
		return "", err
	}
	c.numCacheHits++

	return response, nil
}

func (c *SQLiteCache) SetResponse(ctx context.Context, req llm.InferRequest, response string) error {
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}
	hashedRequest := c.hasher.Sum(requestJSON)

	_, err = c.db.ExecContext(ctx, "INSERT INTO response_cache (request, response) VALUES (?, ?)", hashedRequest, response)
	return err
}

func (c *SQLiteCache) GetEmbedding(ctx context.Context, modelConfig string, input string) ([]float32, error) {
	c.numRequests++
	var embeddingBlob []byte
	err := c.db.QueryRowContext(ctx, "SELECT embedding FROM embedding_cache WHERE model_config = ? AND input_string = ?", modelConfig, input).Scan(&embeddingBlob)
	if err != nil {
		return nil, err
	}
	c.numCacheHits++

	var embedding []float32
	err = json.Unmarshal(embeddingBlob, &embedding)
	if err != nil {
		return nil, err
	}

	return embedding, nil
}

func (c *SQLiteCache) SetEmbedding(ctx context.Context, modelConfig string, input string, embedding []float32) error {
	embeddingBlob, err := json.Marshal(embedding)
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx, "INSERT OR REPLACE INTO embedding_cache (model_config, input_string, embedding) VALUES (?, ?, ?)", modelConfig, input, embeddingBlob)
	return err
}

func (c *SQLiteCache) Close() error {
	return c.db.Close()
}

func (c *SQLiteCache) GetStats() cache.CacheStats {
	return cache.CacheStats{
		NumRequests:  c.numRequests,
		NumCacheHits: c.numCacheHits,
	}
}

func initDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS response_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			request BLOB,
			response TEXT
		);
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

	// Set to WAL mode for better performance
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	return err
}

// Ensure SQLiteCache implements the Cache interface
var _ cache.Cache = (*SQLiteCache)(nil)
