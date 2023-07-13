package gollum_test

import (
	"context"
	"crypto/rand"
	"fmt"
	mathrand "math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stillmatic/gollum"
	"github.com/stretchr/testify/assert"
)

func TestCompressedVectorStore(t *testing.T) {
	vs := gollum.NewStdGzipVectorStore()
	t.Run("implements interface", func(t *testing.T) {
		var vs2 gollum.VectorStore
		vs2 = vs
		assert.NotNil(t, vs2)
	})
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

func BenchmarkCompressedVectorStore(b *testing.B) {
	ctx := context.Background()
	// Test different sizes
	sizes := []int{10, 100, 1000}
	// note that runtime doesn't really depnd on K -
	ks := []int{1, 5, 25, 100}
	// benchmark inserts
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Insert-%d", size), func(b *testing.B) {
			// Create vector store using live compression
			vs := gollum.NewStdGzipVectorStore()
			docs := make([]gollum.Document, size)
			// Generate synthetic docs
			for i := range docs {
				docs[i] = syntheticDoc()
			}
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				// Insert docs
				for _, doc := range docs {
					vs.Insert(ctx, doc)
				}
			}
		})
	}
	// benchmark queries
	for _, size := range sizes {
		for _, k := range ks {
			if k <= size {
				b.Run(fmt.Sprintf("Query-%d-%d", size, k), func(b *testing.B) {
					// Create vector store and insert docs
					vs := gollum.NewStdGzipVectorStore()
					docs := make([]gollum.Document, size)
					for i := range docs {
						docs[i] = syntheticDoc()
						vs.Insert(ctx, docs[i])
					}
					b.ReportAllocs()
					b.ResetTimer()
					// Create query
					query := syntheticQuery(k)
					for n := 0; n < b.N; n++ {
						vs.Query(ctx, query)
					}
				})
			}
		}
	}

}

// Helper functions
func syntheticString() string {
	// Random length between 8 and 64
	randLength := mathrand.Intn(64-8+1) + 8

	// Generate random bytes
	randBytes := make([]byte, randLength)
	rand.Read(randBytes)

	// Format as hex string
	return fmt.Sprintf("%x", randBytes)
}

func syntheticDoc() gollum.Document {
	return gollum.NewDocumentFromString(syntheticString())
}

// syntheticQuery return query request with random embedding
func syntheticQuery(k int) gollum.QueryRequest {
	return gollum.QueryRequest{
		Query: syntheticString(),
		K:     k,
	}
}
