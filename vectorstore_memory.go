package gollum

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/viterin/vek/vek32"
	"gocloud.dev/blob"
)

// MemoryVectorStore embeds documents on insert and stores them in memory
type MemoryVectorStore struct {
	Documents []Document
	LLM       Embedder
}

func NewMemoryVectorStore(llm Embedder) *MemoryVectorStore {
	return &MemoryVectorStore{
		Documents: make([]Document, 0),
		LLM:       llm,
	}
}

func (m *MemoryVectorStore) Insert(ctx context.Context, d Document) error {
	// replace newlines with spaces and strip whitespace, per OpenAI's recommendation
	if d.Embedding == nil {
		cleanText := strings.ReplaceAll(d.Content, "\n", " ")
		cleanText = strings.TrimSpace(cleanText)

		embedding, err := m.LLM.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Input: []string{cleanText},
			// TODO: make this configurable -- may require forking the base library, this expects an enum
			Model: openai.AdaEmbeddingV2,
		})
		if err != nil {
			return errors.Wrap(err, "failed to create embedding")
		}
		d.Embedding = embedding.Data[0].Embedding
	}

	m.Documents = append(m.Documents, d)
	return nil
}

func (m *MemoryVectorStore) Persist(ctx context.Context, bucket *blob.Bucket, path string) error {
	// save documents to disk
	data, err := json.Marshal(m.Documents)
	if err != nil {
		return errors.Wrap(err, "failed to marshal documents to JSON")
	}
	err = bucket.WriteAll(ctx, path, data, nil)
	if err != nil {
		return errors.Wrap(err, "failed to write documents to file")
	}
	return nil
}

func NewMemoryVectorStoreFromDisk(ctx context.Context, bucket *blob.Bucket, path string, llm Embedder) (*MemoryVectorStore, error) {
	data, err := bucket.ReadAll(ctx, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}
	var documents []Document
	err = json.Unmarshal(data, &documents)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON")
	}
	return &MemoryVectorStore{
		Documents: documents,
		LLM:       llm,
	}, nil
}

func (m *MemoryVectorStore) Query(ctx context.Context, qb QueryRequest) ([]Document, error) {
	if len(m.Documents) == 0 {
		return nil, errors.New("no documents in store")
	}
	if len(qb.EmbeddingStrings) > 0 {
		// concatenate strings and set query
		qb.Query = strings.Join(qb.EmbeddingStrings, " ")
	}
	if len(qb.EmbeddingFloats) == 0 {
		// create embedding
		embedding, err := m.LLM.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Input: []string{qb.Query},
			// TODO: make this configurable
			Model: openai.AdaEmbeddingV2,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create embedding")
		}
		qb.EmbeddingFloats = embedding.Data[0].Embedding
	}
	scores := Heap{}
	scores.Init()
	k := qb.K

	for _, doc := range m.Documents {
		score := vek32.CosineSimilarity(qb.EmbeddingFloats, doc.Embedding)
		ns := nodeSimilarity{
			Document:   &doc,
			Similarity: score,
		}
		// maintain a max-heap of size k
		scores.Push(ns)
		if scores.Len() > k {
			scores.Pop()
		}
	}

	result := make([]Document, k)
	for i := 0; i < k; i++ {
		result[k-i-1] = *(scores.Pop().Document)
	}
	return result, nil
}

// RetrieveAll returns all documents
func (m *MemoryVectorStore) RetrieveAll(ctx context.Context) ([]Document, error) {
	return m.Documents, nil
}
