package cached_test

import (
	"context"
	"testing"

	"github.com/stillmatic/gollum/packages/llm"
	mock_llm "github.com/stillmatic/gollum/packages/llm/internal/mocks"
	"github.com/stillmatic/gollum/packages/llm/providers/cached"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCachedProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Run("responder", func(t *testing.T) {
		mockProvider := mock_llm.NewMockResponder(ctrl)
		ctx := context.Background()
		req := llm.InferRequest{
			Messages: []llm.InferMessage{
				{Content: "hello world",
					Role: "user",
				},
			},
			ModelConfig: llm.ModelConfig{
				ModelName:    "fake_model",
				ProviderType: llm.ProviderAnthropic,
			},
		}

		mockProvider.EXPECT().GenerateResponse(ctx, req).Return("hello user", nil)

		cachedProvider, err := cached.NewLocalCachedResponder(mockProvider, ":memory:")
		assert.NoError(t, err)
		resp, err := cachedProvider.GenerateResponse(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, "hello user", resp)

		cs := cachedProvider.GetCacheStats()
		assert.Equal(t, 1, cs.NumRequests)
		assert.Equal(t, 0, cs.NumCacheHits)

		resp, err = cachedProvider.GenerateResponse(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, "hello user", resp)

		cs = cachedProvider.GetCacheStats()
		assert.Equal(t, 2, cs.NumRequests)
		assert.Equal(t, 1, cs.NumCacheHits)
	})

	t.Run("embedder", func(t *testing.T) {
		mockProvider := mock_llm.NewMockEmbedder(ctrl)
		ctx := context.Background()
		req := llm.EmbedRequest{
			Input: []string{"abc"},
			ModelConfig: llm.ModelConfig{
				ModelName:    "fake_model",
				ProviderType: llm.ProviderAnthropic,
			},
		}

		// we call the function twice but it returns the cached value second time, so
		// the provider should only be called once
		mockProvider.EXPECT().GenerateEmbedding(ctx, req).Return(&llm.EmbeddingResponse{
			Data: []llm.Embedding{{Values: []float32{1.0, 2.0, 3.0}}}}, nil).Times(1)

		cachedProvider, err := cached.NewLocalCachedEmbedder(mockProvider, ":memory:")
		assert.NoError(t, err)
		resp, err := cachedProvider.GenerateEmbedding(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, []float32{1.0, 2.0, 3.0}, resp.Data[0].Values)

		cs := cachedProvider.GetCacheStats()
		assert.Equal(t, 1, cs.NumRequests)
		assert.Equal(t, 0, cs.NumCacheHits)

		resp, err = cachedProvider.GenerateEmbedding(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, []float32{1.0, 2.0, 3.0}, resp.Data[0].Values)

		cs = cachedProvider.GetCacheStats()
		assert.Equal(t, 2, cs.NumRequests)
		assert.Equal(t, 1, cs.NumCacheHits)

		req = llm.EmbedRequest{
			Input: []string{"abc", "def"},
			ModelConfig: llm.ModelConfig{
				ModelName:    "fake_model",
				ProviderType: llm.ProviderAnthropic,
			},
		}
		// this should be the request to the provider since we don't have the second embedding cached
		cachedReq := llm.EmbedRequest{
			Input: []string{"def"},
			ModelConfig: llm.ModelConfig{
				ModelName:    "fake_model",
				ProviderType: llm.ProviderAnthropic,
			},
		}
		// and the provider only returns one
		mockProvider.EXPECT().GenerateEmbedding(ctx, cachedReq).Return(&llm.EmbeddingResponse{
			Data: []llm.Embedding{{Values: []float32{2.0, 3.0, 4.0}}}}, nil).Times(1)

		//but we should expect to get both embeddings back
		resp, err = cachedProvider.GenerateEmbedding(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, []float32{1.0, 2.0, 3.0}, resp.Data[0].Values)
		assert.Equal(t, []float32{2.0, 3.0, 4.0}, resp.Data[1].Values)

		cs = cachedProvider.GetCacheStats()
		assert.Equal(t, 4, cs.NumRequests)
		assert.Equal(t, 2, cs.NumCacheHits)
	})
}
