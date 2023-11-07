package hyde_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stillmatic/gollum"
	"github.com/stillmatic/gollum/hyde"
	mock_gollum "github.com/stillmatic/gollum/internal/mocks"
	"github.com/stillmatic/gollum/internal/testutil"
	"github.com/stillmatic/gollum/vectorstore"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHyde(t *testing.T) {
	ctrl := gomock.NewController(t)
	embedder := mock_gollum.NewMockEmbedder(ctrl)
	completer := mock_gollum.NewMockChatCompleter(ctrl)
	prompter := hyde.NewZeroShotPrompter(
		"Roleplay as a character. Write a short biographical answer to the question.\nQ: %s\nA:",
	)
	generator := hyde.NewLLMGenerator(completer)
	encoder := hyde.NewLLMEncoder(embedder)
	vs := vectorstore.NewMemoryVectorStore(embedder)
	for i := range make([]int, 10) {
		embedder.EXPECT().CreateEmbeddings(context.Background(), gomock.Any()).Return(testutil.GetRandomEmbeddingResponse(1, 1536), nil)
		vs.Insert(context.Background(), gollum.NewDocumentFromString(fmt.Sprintf("hey %d", i)))
	}
	assert.Equal(t, 10, len(vs.Documents))
	searcher := hyde.NewVectorSearcher(
		vs,
	)

	hyde := hyde.NewHyde(prompter, generator, encoder, searcher)
	t.Run("prompter", func(t *testing.T) {
		prompt := prompter.BuildPrompt(context.Background(), "What is your name?")
		assert.Equal(t, "Roleplay as a character. Write a short biographical answer to the question.\nQ: What is your name?\nA:", prompt)
	})

	t.Run("generator", func(t *testing.T) {
		ctx := context.Background()
		k := 10
		completer.EXPECT().CreateChatCompletion(ctx, gomock.Any()).Return(testutil.GetRandomChatCompletionResponse(k), nil)
		res, err := generator.Generate(ctx, "What is your name?", k)
		assert.NoError(t, err)
		assert.Equal(t, 10, len(res))
	})

	t.Run("encoder", func(t *testing.T) {
		ctx := context.Background()
		embedder.EXPECT().CreateEmbeddings(ctx, gomock.Any()).Return(testutil.GetRandomEmbeddingResponse(1, 1536), nil)
		res, err := encoder.Encode(ctx, "What is your name?")
		assert.NoError(t, err)
		assert.Equal(t, 1536, len(res))

		embedder.EXPECT().CreateEmbeddings(ctx, gomock.Any()).Return(testutil.GetRandomEmbeddingResponse(2, 1536), nil)
		res2, err := encoder.EncodeBatch(ctx, []string{"What is your name?", "What is your quest?"})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(res2))
		assert.Equal(t, 1536, len(res2[0]))
	})

	t.Run("e2e", func(t *testing.T) {
		ctx := context.Background()
		completer.EXPECT().CreateChatCompletion(ctx, gomock.Any()).Return(testutil.GetRandomChatCompletionResponse(3), nil)
		embedder.EXPECT().CreateEmbeddings(ctx, gomock.Any()).Return(testutil.GetRandomEmbeddingResponse(4, 1536), nil)
		res, err := hyde.SearchEndToEnd(ctx, "What is your name?", 3)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(res))
	})
}

func BenchmarkHyde(b *testing.B) {
	prompter := hyde.NewZeroShotPrompter(
		"Roleplay as a character. Write a short biographical answer to the question.\nQ: %s\nA:",
	)
	ctrl := gomock.NewController(b)
	embedder := mock_gollum.NewMockEmbedder(ctrl)
	completer := mock_gollum.NewMockChatCompleter(ctrl)
	generator := hyde.NewLLMGenerator(completer)
	encoder := hyde.NewLLMEncoder(embedder)
	vs := vectorstore.NewMemoryVectorStore(embedder)
	searcher := hyde.NewVectorSearcher(
		vs,
	)
	hyde := hyde.NewHyde(prompter, generator, encoder, searcher)
	docNums := []int{10, 100, 1000, 10000, 100_000, 1_000_000}
	for _, docNum := range docNums {
		b.Run(fmt.Sprintf("docs=%v", docNum), func(b *testing.B) {
			for _ = range make([]int, docNum) {
				embedder.EXPECT().CreateEmbeddings(gomock.Any(), gomock.Any()).Return(testutil.GetRandomEmbeddingResponse(1, 1536), nil)
				vs.Insert(context.Background(), gollum.NewDocumentFromString("hey"))
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				k := 8
				completer.EXPECT().CreateChatCompletion(context.Background(), gomock.Any()).Return(testutil.GetRandomChatCompletionResponse(k), nil)
				embedder.EXPECT().CreateEmbeddings(context.Background(), gomock.Any()).Return(testutil.GetRandomEmbeddingResponse(k+1, 1536), nil)
				_, err := hyde.SearchEndToEnd(context.Background(), "What is your name?", k)
				assert.NoError(b, err)
			}
		})
	}
}
