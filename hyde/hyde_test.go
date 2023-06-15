package hyde_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sashabaranov/go-openai"
	"github.com/stillmatic/gollum"
	"github.com/stillmatic/gollum/hyde"
	mock_gollum "github.com/stillmatic/gollum/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func getRandomEmbedding(n int) []float32 {
	vec := make([]float32, n)
	for i := range vec {
		vec[i] = rand.Float32()
	}
	return vec
}

func getRandomEmbeddingResponse(n int, dim int) openai.EmbeddingResponse {
	data := make([]openai.Embedding, n)
	for i := range data {
		data[i] = openai.Embedding{
			Embedding: getRandomEmbedding(dim),
		}
	}
	resp := openai.EmbeddingResponse{
		Data: data,
	}
	return resp
}

func getRandomChatCompletionResponse(n int) openai.ChatCompletionResponse {
	choices := make([]openai.ChatCompletionChoice, n)
	for i := range choices {
		choices[i] = openai.ChatCompletionChoice{
			Message: openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("test? %d", i),
			},
		}
	}
	return openai.ChatCompletionResponse{
		Choices: choices,
	}
}

func TestHyde(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLLM := mock_gollum.NewMockLLM(ctrl)
	prompter := hyde.NewZeroShotPrompter(
		"Roleplay as a character. Write a short biographical answer to the question.\nQ: %s\nA:",
	)
	generator := hyde.NewLLMGenerator(
		mockLLM,
	)
	encoder := hyde.NewLLMEncoder(
		mockLLM,
	)
	vs := gollum.NewMemoryVectorStore(
		mockLLM,
	)
	for i := range make([]int, 10) {
		mockLLM.EXPECT().CreateEmbeddings(context.Background(), gomock.Any()).Return(getRandomEmbeddingResponse(1, 1536), nil)
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
		mockLLM.EXPECT().CreateChatCompletion(ctx, gomock.Any()).Return(getRandomChatCompletionResponse(k), nil)
		res, err := generator.Generate(ctx, "What is your name?", k)
		assert.NoError(t, err)
		assert.Equal(t, 10, len(res))
	})

	t.Run("encoder", func(t *testing.T) {
		ctx := context.Background()
		mockLLM.EXPECT().CreateEmbeddings(ctx, gomock.Any()).Return(getRandomEmbeddingResponse(1, 1536), nil)
		res, err := encoder.Encode(ctx, "What is your name?")
		assert.NoError(t, err)
		assert.Equal(t, 1536, len(res))

		mockLLM.EXPECT().CreateEmbeddings(ctx, gomock.Any()).Return(getRandomEmbeddingResponse(2, 1536), nil)
		res2, err := encoder.EncodeBatch(ctx, []string{"What is your name?", "What is your quest?"})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(res2))
		assert.Equal(t, 1536, len(res2[0]))
	})

	t.Run("e2e", func(t *testing.T) {
		ctx := context.Background()
		mockLLM.EXPECT().CreateChatCompletion(ctx, gomock.Any()).Return(getRandomChatCompletionResponse(3), nil)
		mockLLM.EXPECT().CreateEmbeddings(ctx, gomock.Any()).Return(getRandomEmbeddingResponse(4, 1536), nil)
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
	mockLLM := mock_gollum.NewMockLLM(ctrl)
	generator := hyde.NewLLMGenerator(mockLLM)
	encoder := hyde.NewLLMEncoder(mockLLM)
	vs := gollum.NewMemoryVectorStore(mockLLM)
	searcher := hyde.NewVectorSearcher(
		vs,
	)
	hyde := hyde.NewHyde(prompter, generator, encoder, searcher)
	docNums := []int{10, 100, 1000, 10000, 100_000, 1_000_000}
	for _, docNum := range docNums {
		b.Run(fmt.Sprintf("docs=%v", docNum), func(b *testing.B) {
			for _ = range make([]int, docNum) {
				mockLLM.EXPECT().CreateEmbeddings(gomock.Any(), gomock.Any()).Return(getRandomEmbeddingResponse(1, 1536), nil)
				vs.Insert(context.Background(), gollum.NewDocumentFromString("hey"))
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				k := 8
				mockLLM.EXPECT().CreateChatCompletion(context.Background(), gomock.Any()).Return(getRandomChatCompletionResponse(k), nil)
				mockLLM.EXPECT().CreateEmbeddings(context.Background(), gomock.Any()).Return(getRandomEmbeddingResponse(k+1, 1536), nil)
				_, err := hyde.SearchEndToEnd(context.Background(), "What is your name?", k)
				assert.NoError(b, err)
			}
		})
	}
}
