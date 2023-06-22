package gollum

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	openai "github.com/sashabaranov/go-openai"
	"github.com/viterin/vek/vek32"
	"gocloud.dev/blob"
)

// QueryRequest is a struct that contains the query and optional query strings or embeddings
type QueryRequest struct {
	// Query is the text to query
	Query string
	// EmbeddingStrings is a list of strings to concatenate and embed instead of Query
	EmbeddingStrings []string
	// EmbeddingFloats is a query vector to use instead of Query
	EmbeddingFloats []float32
	// K is the number of results to return
	K int
}

type VectorStore interface {
	Insert(context.Context, Document) error
	Query(ctx context.Context, qb QueryRequest) ([]Document, error)
	RetrieveAll(ctx context.Context) ([]Document, error)
}

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

type nodeSimilarity struct {
	Document   Document
	Similarity float32
}

// Heap is a custom heap implementation, to avoid interface{} conversion.
// I _think_ theoretically that a memory arena would be useful here, but that feels a bit beyond the pale, even for me.
// In benchmarking, we see that allocations are limited by scale according to k --
// since K is known, we should be able to allocate a fixed-size arena and use that.
// That being said... let's revisit in the future :)
type Heap []nodeSimilarity

func (h Heap) Init() {
	for i := (len(h) - 1) / 2; i >= 0; i-- {
		h.down(i)
	}
}

func (h Heap) down(u int) {
	v := u
	if 2*u+1 < len(h) && h[2*u+1].Similarity < h[v].Similarity {
		v = 2*u + 1
	}
	if 2*u+2 < len(h) && h[2*u+2].Similarity < h[v].Similarity {
		v = 2*u + 2
	}
	if v != u {
		h[v], h[u] = h[u], h[v]
		h.down(v)
	}
}

func (h Heap) up(u int) {
	for u != 0 && h[(u-1)/2].Similarity > h[u].Similarity {
		h[(u-1)/2], h[u] = h[u], h[(u-1)/2]
		u = (u - 1) / 2
	}
}

func (h *Heap) Push(e nodeSimilarity) {
	*h = append(*h, e)
	h.up(len(*h) - 1)
}

func (h *Heap) Pop() nodeSimilarity {
	x := (*h)[0]
	n := len(*h)
	(*h)[0], (*h)[n-1] = (*h)[n-1], (*h)[0]
	*h = (*h)[:n-1]
	h.down(0)
	return x
}

func (h *Heap) Len() int {
	return len(*h)
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
			Document:   doc,
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
		result[k-i-1] = scores.Pop().Document
	}
	return result, nil
}

// RetrieveAll returns all documents
func (m *MemoryVectorStore) RetrieveAll(ctx context.Context) ([]Document, error) {
	return m.Documents, nil
}
