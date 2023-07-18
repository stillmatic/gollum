package gollum_test

import (
	"context"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stillmatic/gollum"
	mock_gollum "github.com/stillmatic/gollum/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type testInput struct {
	Topic       string   `json:"topic" jsonschema:"required" jsonschema_description:"The topic of the conversation"`
	RandomWords []string `json:"random_words" jsonschema:"required" jsonschema_description:"Random words to prime the conversation"`
}

func TestDummyDispatcher(t *testing.T) {
	d := gollum.NewDummyDispatcher[testInput]()

	output, err := d.Prompt(context.Background(), "Hello")

	assert.NoError(t, err)
	assert.Equal(t, testInput{}, output)
}

func TestOpenAIDispatcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	completer := mock_gollum.NewMockChatCompleter(ctrl)
	d := gollum.NewOpenAIDispatcher[testInput]("Random Conversation", "Given a topic, return random words", completer)

	ctx := context.Background()
	queryStr := "Tell me about dinosaurs"
	inpStr := `{"topic": "dinosaurs", "random_words": ["dinosaur", "fossil", "extinct"]}`
	completer.EXPECT().CreateChatCompletion(gomock.Any(), gomock.Any()).Return(openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: "Hello there!",
					FunctionCall: &openai.FunctionCall{
						Name:      "Random Conversation",
						Arguments: inpStr,
					},
				},
			},
		},
	}, nil)

	output, err := d.Prompt(ctx, queryStr)
	assert.NoError(t, err)
	expected := testInput{
		Topic:       "dinosaurs",
		RandomWords: []string{"dinosaur", "fossil", "extinct"},
	}
	assert.Equal(t, expected, output)
}
