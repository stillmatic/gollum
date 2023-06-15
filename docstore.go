package gollum

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"gocloud.dev/blob"
)

type DocStore interface {
	Insert(context.Context, Document) error
	Retrieve(ctx context.Context, id string, k int) ([]Document, error)
}

// MemoryDocStore is a simple in-memory document store.
// It's functionally a hashmap / inverted-index.
type MemoryDocStore struct {
	Documents map[string]Document
}

func NewMemoryDocStore() *MemoryDocStore {
	return &MemoryDocStore{
		Documents: make(map[string]Document),
	}
}

func NewMemoryDocStoreFromDisk(ctx context.Context, bucket *blob.Bucket, path string) (*MemoryDocStore, error) {
	// load documents from disk
	data, err := bucket.ReadAll(ctx, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read documents from disk")
	}
	var nodes map[string]Document
	err = json.Unmarshal(data, &nodes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal documents from JSON")
	}
	return &MemoryDocStore{
		Documents: nodes,
	}, nil
}

// Insert adds a node to the document store. It overwrites duplicates.
func (m *MemoryDocStore) Insert(ctx context.Context, d Document) error {
	m.Documents[d.ID] = d
	return nil
}

// Retrieve returns a node from the document store matching an ID.
func (m *MemoryDocStore) Retrieve(ctx context.Context, id string) (Document, error) {
	v, ok := m.Documents[id]
	if !ok {
		return Document{}, errors.New("document not found")
	}
	return v, nil
}

// Persist saves the document store to disk.
func (m *MemoryDocStore) Persist(ctx context.Context, bucket *blob.Bucket, path string) error {
	// save documents to disk
	data, err := json.Marshal(m.Documents)
	if err != nil {
		return errors.Wrap(err, "failed to marshal documents to JSON")
	}
	err = bucket.WriteAll(ctx, path, data, nil)
	if err != nil {
		return errors.Wrap(err, "failed to write documents to disk")
	}
	return nil
}
