package tools

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stillmatic/gollum"
)

type Retriever struct {
	vs   gollum.VectorStore
	docs []gollum.Document
}

func (r *Retriever) Name() string {
	return "Retriever"
}

func (r *Retriever) Description() string {
	return "Retrieves a document from a vector store"
}

// NewRetriever returns a new Retriever tool reading from the given vector store.
func NewRetriever(vs gollum.VectorStore) *Retriever {
	return &Retriever{
		vs:   vs,
		docs: make([]gollum.Document, 0),
	}
}

func (r *Retriever) Run(ctx context.Context, input interface{}) (interface{}, error) {
	doc, ok := input.(gollum.QueryRequest)
	if !ok {
		return nil, errors.New("invalid input")
	}
	return r.vs.Query(ctx, doc)
}

func (r *Retriever) Write(ctx context.Context, input interface{}) error {
	doc, ok := input.(gollum.QueryRequest)
	if !ok {
		return errors.New("invalid input")
	}
	defer func() {
		resp, err := r.vs.Query(ctx, doc)
		if err != nil {
			return
		}
		r.docs = resp
	}()
	return nil
}
