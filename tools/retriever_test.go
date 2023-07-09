package tools_test

import (
	"context"
	"testing"

	"github.com/stillmatic/gollum"
	mock_gollum "github.com/stillmatic/gollum/internal/mocks"
	"github.com/stillmatic/gollum/tools"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRetriever(t *testing.T) {
	ctrl := gomock.NewController(t)
	embedder := mock_gollum.NewMockEmbedder(ctrl)
	vs := gollum.NewMemoryVectorStore(embedder)
	retriever := tools.NewRetriever(vs)

	t.Run("simple", func(t *testing.T) {
		ctx := context.Background()
		input := gollum.QueryRequest{
			Query: "Apple",
		}
		output, err := retriever.Run(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, "Apple", output)
	})
}
