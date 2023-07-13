package gollum_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stillmatic/gollum"
	"github.com/stretchr/testify/assert"
)

func TestCompressedVectorStore(t *testing.T) {
	var vs gollum.VectorStore

	vs = gollum.NewGzipVectorStore()
	ctx := context.Background()
	testStrings := []string{
		"apple",
		"orange",
		"basketball",
		"football",
		"banana",
	}
	t.Run("testInsert", func(t *testing.T) {
		for _, str := range testStrings {
			vs.Insert(ctx, gollum.Document{
				ID:      uuid.NewString(),
				Content: str,
			})
		}
		docs, err := vs.RetrieveAll(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 5, len(docs))
	})
	t.Run("correctness", func(t *testing.T) {
		for _, str := range testStrings {
			vs.Insert(ctx, gollum.Document{
				ID:      uuid.NewString(),
				Content: str,
			})
		}
		docs, err := vs.Query(ctx, gollum.QueryRequest{
			Query: "tennis ball",
			K:     5,
		})
		assert.NoError(t, err)
		assert.Equal(t, 5, len(docs))
		assert.Equal(t, "football", docs[0].Content)
		assert.Equal(t, "basketball", docs[1].Content)
	})
}
