package gollum

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type Dispatcher[T any] interface {
	Prompt(ctx context.Context, prompt string) (interface{}, error)
}

type OpenAIDispatcher[T any] struct {
	client *openai.Client
	fi     openai.FunctionDefinition
	parser Parser[T]
}

func NewOpenAIDispatcher[T any](name, description string, client *openai.Client) *OpenAIDispatcher[T] {
	var t T
	fi := StructToJsonSchema(name, description, t)
	parser := NewJSONParserGeneric[T](true)
	return &OpenAIDispatcher[T]{
		client: client,
		fi:     openai.FunctionDefinition(fi),
		parser: parser,
	}
}

func (d *OpenAIDispatcher[T]) Prompt(ctx context.Context, prompt string) (T, error) {
	var output T
	resp, err := d.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		// TODO: configure this
		Model: openai.GPT3Dot5Turbo0613,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
		},
		Functions: []openai.FunctionDefinition{d.fi},
		// TODO: configure this
		Temperature: 0.0,
		MaxTokens:   1000,
	})
	if err != nil {
		return output, err
	}

	output, err = d.parser.Parse(ctx, []byte(resp.Choices[0].Message.FunctionCall.Arguments))
	if err != nil {
		return output, err
	}

	return output, nil
}
