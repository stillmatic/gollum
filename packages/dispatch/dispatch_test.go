package dispatch_test

import (
	"context"
	"os"
	"testing"
	"text/template"

	"github.com/sashabaranov/go-openai"
	mock_gollum "github.com/stillmatic/gollum/internal/mocks"
	"github.com/stillmatic/gollum/packages/dispatch"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type testInput struct {
	Topic       string   `json:"topic" jsonschema:"required" jsonschema_description:"The topic of the conversation"`
	RandomWords []string `json:"random_words" jsonschema:"required" jsonschema_description:"Random words to prime the conversation"`
}

type templateInput struct {
	Topic string
}

type wordCountOutput struct {
	Count int `json:"count" jsonschema:"required" jsonschema_description:"The number of words in the sentence"`
}

func TestDummyDispatcher(t *testing.T) {
	d := dispatch.NewDummyDispatcher[testInput]()

	t.Run("prompt", func(t *testing.T) {
		output, err := d.Prompt(context.Background(), "Talk to me about Dinosaurs")

		assert.NoError(t, err)
		assert.Equal(t, testInput{}, output)
	})
	t.Run("promptTemplate", func(t *testing.T) {
		te, err := template.New("").Parse("Talk to me about {{.Topic}}")
		assert.NoError(t, err)
		tempInp := templateInput{
			Topic: "Dinosaurs",
		}

		output, err := d.PromptTemplate(context.Background(), te, tempInp)
		assert.NoError(t, err)
		assert.Equal(t, testInput{}, output)
	})
}

func TestOpenAIDispatcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	completer := mock_gollum.NewMockChatCompleter(ctrl)
	systemPrompt := "When prompted, use the tool."
	d := dispatch.NewOpenAIDispatcher[testInput]("random_conversation", "Given a topic, return random words", systemPrompt, completer, nil)

	ctx := context.Background()
	expected := testInput{
		Topic:       "dinosaurs",
		RandomWords: []string{"dinosaur", "fossil", "extinct"},
	}
	inpStr := `{"topic": "dinosaurs", "random_words": ["dinosaur", "fossil", "extinct"]}`

	fi := openai.FunctionDefinition(dispatch.StructToJsonSchema("random_conversation", "Given a topic, return random words", testInput{}))
	ti := openai.Tool{Type: "function", Function: &fi}
	expectedRequest := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo1106,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Tell me about dinosaurs",
			},
		},
		Tools: []openai.Tool{ti},
		ToolChoice: openai.ToolChoice{
			Type: "function",
			Function: openai.ToolFunction{
				Name: "random_conversation",
			}},
		MaxTokens:   512,
		Temperature: 0.0,
	}

	t.Run("prompt", func(t *testing.T) {
		queryStr := "Tell me about dinosaurs"
		completer.EXPECT().CreateChatCompletion(gomock.Any(), expectedRequest).Return(openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleSystem,
						Content: "Hello there!",
						ToolCalls: []openai.ToolCall{
							{
								Type: "function",
								Function: openai.FunctionCall{
									Name:      "random_conversation",
									Arguments: inpStr,
								},
							}},
					},
				},
			},
		}, nil)

		output, err := d.Prompt(ctx, queryStr)
		assert.NoError(t, err)

		assert.Equal(t, expected, output)
	})
	t.Run("promptTemplate", func(t *testing.T) {
		completer.EXPECT().CreateChatCompletion(gomock.Any(), expectedRequest).Return(openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleSystem,
						Content: "Hello there!",
						ToolCalls: []openai.ToolCall{
							{
								Type: "function",
								Function: openai.FunctionCall{
									Name:      "random_conversation",
									Arguments: inpStr,
								},
							}},
					},
				},
			},
		}, nil)

		te, err := template.New("").Parse("Tell me about {{.Topic}}")
		assert.NoError(t, err)

		output, err := d.PromptTemplate(ctx, te, templateInput{
			Topic: "dinosaurs",
		})
		assert.NoError(t, err)
		assert.Equal(t, expected, output)
	})
}

func TestDispatchIntegration(t *testing.T) {
	t.Skip("Skipping integration test")
	completer := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	systemPrompt := "When prompted, use the tool on the user's input."
	d := dispatch.NewOpenAIDispatcher[wordCountOutput]("wordCounter", "count the number of words in a sentence", systemPrompt, completer, nil)
	output, err := d.Prompt(context.Background(), "I like dinosaurs")
	assert.NoError(t, err)
	assert.Equal(t, 3, output.Count)
}
