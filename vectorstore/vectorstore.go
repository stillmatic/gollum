package vectorstore

import (
	"context"

	"github.com/stillmatic/gollum"
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
	Insert(context.Context, gollum.Document) error
	Query(ctx context.Context, qb QueryRequest) ([]*gollum.Document, error)
	RetrieveAll(ctx context.Context) ([]gollum.Document, error)
}

type NodeSimilarity struct {
	Document   *gollum.Document
	Similarity float32
}

// Heap is a custom heap implementation, to avoid interface{} conversion.
// I _think_ theoretically that a memory arena would be useful here, but that feels a bit beyond the pale, even for me.
// In benchmarking, we see that allocations are limited by scale according to k --
// since K is known, we should be able to allocate a fixed-size arena and use that.
// That being said... let's revisit in the future :)
type Heap []NodeSimilarity

func (h *Heap) Init(k int) {
	*h = make(Heap, 0, k)
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

func (h *Heap) Push(e NodeSimilarity) {
	*h = append(*h, e)
	h.up(len(*h) - 1)
}

func (h *Heap) Pop() NodeSimilarity {
	x := (*h)[0]
	n := len(*h)
	(*h)[0], (*h)[n-1] = (*h)[n-1], (*h)[0]
	*h = (*h)[:n-1]
	h.down(0)
	return x
}

func (h Heap) Less(i, j int) bool {
	return h[i].Similarity < h[j].Similarity
}

func (h Heap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *Heap) Len() int {
	return len(*h)
}
