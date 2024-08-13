package vectorstore_test

import (
	"context"
	"fmt"
	vectorstore2 "github.com/stillmatic/gollum/packages/vectorstore"
	"math/rand"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stillmatic/gollum"
	mock_gollum "github.com/stillmatic/gollum/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gocloud.dev/blob/fileblob"
)

func getRandomEmbedding(n int) []float32 {
	vec := make([]float32, n)
	for i := range vec {
		vec[i] = rand.Float32()
	}
	return vec
}

// setup with godotenv load
func initialize(tb testing.TB) (*mock_gollum.MockEmbedder, *vectorstore2.MemoryVectorStore) {
	tb.Helper()

	ctrl := gomock.NewController(tb)
	oai := mock_gollum.NewMockEmbedder(ctrl)
	ctx := context.Background()
	bucket, err := fileblob.OpenBucket("testdata", nil)
	assert.NoError(tb, err)
	mvs, err := vectorstore2.NewMemoryVectorStoreFromDisk(ctx, bucket, "simple_store.json", oai)
	if err != nil {
		fmt.Println(err)
		mvs = vectorstore2.NewMemoryVectorStore(oai)
		testStrs := []string{"Apple", "Orange", "Basketball"}
		for i, s := range testStrs {
			mv := gollum.NewDocumentFromString(s)
			expectedReq := openai.EmbeddingRequest{
				Input: []string{s},
				Model: openai.AdaEmbeddingV2,
			}
			val := float64(i) + 0.1
			expectedResp := openai.EmbeddingResponse{
				Data: []openai.Embedding{{Embedding: []float32{float32(0.1), float32(val), float32(val)}}},
			}
			oai.EXPECT().CreateEmbeddings(ctx, expectedReq).Return(expectedResp, nil)
			err := mvs.Insert(ctx, mv)
			assert.NoError(tb, err)
		}
		err := mvs.Persist(ctx, bucket, "simple_store.json")
		assert.NoError(tb, err)
	}
	return oai, mvs
}

// TestRetrieval tests inserting embeddings and retrieving them
func TestMemoryVectorStore(t *testing.T) {
	mockllm, mvs := initialize(t)
	ctx := context.Background()
	t.Run("LoadFromDisk", func(t *testing.T) {
		t.Log(mvs.Documents)
		assert.Equal(t, 3, len(mvs.Documents))
		// check that an ID is in the map
		testStrs := []string{"Apple", "Orange", "Basketball"}
		// test that all the strings are in the documents
		for _, s := range mvs.Documents {
			found := false
			for _, t := range testStrs {
				if s.Content == t {
					found = true
				}
			}
			assert.True(t, found)
		}
	})

	// should return apple and orange first.
	t.Run("QueryWithQuery", func(t *testing.T) {
		k := 2
		qb := vectorstore2.QueryRequest{
			Query: "favorite fruit?",
			K:     k,
		}
		expectedCreateReq := openai.EmbeddingRequest{
			Input: []string{"favorite fruit?"},
			Model: openai.AdaEmbeddingV2,
		}
		expectedCreateResp := openai.EmbeddingResponse{
			Data: []openai.Embedding{
				{
					Embedding: []float32{float32(0.1), float32(0.1), float32(0.1)},
				},
			},
		}
		mockllm.EXPECT().CreateEmbeddings(ctx, expectedCreateReq).Return(expectedCreateResp, nil)
		resp, err := mvs.Query(ctx, qb)
		assert.NoError(t, err)
		assert.Equal(t, k, len(resp))
		assert.Equal(t, "Apple", resp[0].Content)
		assert.Equal(t, "Orange", resp[1].Content)
	})

	// This should return basketball because the embedding str should override the query
	t.Run("QueryWithEmbedding", func(t *testing.T) {
		k := 1
		qb := vectorstore2.QueryRequest{
			Query:            "What is your favorite fruit",
			EmbeddingStrings: []string{"favorite sport?"},
			K:                k,
		}
		expectedCreateReq := openai.EmbeddingRequest{
			Input: []string{"favorite sport?"},
			Model: openai.AdaEmbeddingV2,
		}
		expectedCreateResp := openai.EmbeddingResponse{
			Data: []openai.Embedding{
				{Embedding: []float32{float32(0.1), float32(2.11), float32(2.11)}},
			}}
		mockllm.EXPECT().CreateEmbeddings(ctx, expectedCreateReq).Return(expectedCreateResp, nil)
		resp, err := mvs.Query(ctx, qb)
		assert.NoError(t, err)
		assert.Equal(t, k, len(resp))
		assert.Equal(t, "Basketball", resp[0].Content)
	})
}

type MockEmbedder struct{}

func (m MockEmbedder) CreateEmbeddings(ctx context.Context, req openai.EmbeddingRequest) (openai.EmbeddingResponse, error) {
	resp := openai.EmbeddingResponse{
		Data: []openai.Embedding{
			{Embedding: getRandomEmbedding(1536)},
		},
	}
	return resp, nil
}

func BenchmarkMemoryVectorStore(b *testing.B) {
	llm := mock_gollum.NewMockEmbedder(gomock.NewController(b))
	ctx := context.Background()

	nValues := []int{10, 100, 1_000, 10_000, 100_000, 1_000_000}
	kValues := []int{1, 10, 100}
	dim := 768
	for _, n := range nValues {
		b.Run(fmt.Sprintf("BenchmarkInsert-n=%v", n), func(b *testing.B) {
			mvs := vectorstore2.NewMemoryVectorStore(llm)
			for i := 0; i < b.N; i++ {
				for j := 0; j < n; j++ {
					mv := gollum.Document{
						ID:        fmt.Sprintf("%v", j),
						Content:   "test",
						Embedding: getRandomEmbedding(dim),
					}
					mvs.Insert(ctx, mv)
				}
			}
		})
		for _, k := range kValues {
			if k <= n {
				b.Run(fmt.Sprintf("BenchmarkQuery-n=%v-k=%v", n, k), func(b *testing.B) {
					mvs := vectorstore2.NewMemoryVectorStore(llm)
					for j := 0; j < n; j++ {
						mv := gollum.Document{
							ID:        fmt.Sprintf("%v", j),
							Content:   "test",
							Embedding: getRandomEmbedding(dim),
						}
						mvs.Insert(ctx, mv)
					}
					qb := vectorstore2.QueryRequest{
						EmbeddingFloats: getRandomEmbedding(dim),
						K:               k,
					}
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						_, err := mvs.Query(ctx, qb)
						assert.NoError(b, err)
					}
				})
			}
		}
	}
}

func BenchmarkHeap(b *testing.B) {
	// Create a sample Heap.

	ks := []int{1, 10, 100}

	for _, k := range ks {
		var h vectorstore2.Heap
		h.Init(k)
		b.Run(fmt.Sprintf("BenchmarkHeapPush-k=%v", k), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				doc := &gollum.Document{}
				similarity := rand.Float32()
				ns := vectorstore2.NodeSimilarity{Document: doc, Similarity: similarity}
				h.Push(ns)
				if h.Len() > k {
					h.Pop()
				}
			}
		})
	}
}
