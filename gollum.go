package gollum

import (
	"github.com/google/uuid"
)

type Document struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content,omitempty"`
	Embedding []float32              `json:"embedding,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

func NewDocumentFromString(content string) Document {
	return Document{
		ID:      uuid.New().String(),
		Content: content,
	}
}
