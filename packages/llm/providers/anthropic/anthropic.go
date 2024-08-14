package anthropic

import (
	"context"
	"github.com/stillmatic/gollum/packages/llm"
	"log/slog"

	"github.com/liushuangls/go-anthropic"
	"github.com/pkg/errors"
)

type Provider struct {
	client *anthropic.Client
}

func NewAnthropicProvider(apiKey string) *Provider {
	return &Provider{
		client: anthropic.NewClient(apiKey),
	}
}

func (p *Provider) GenerateResponse(ctx context.Context, req llm.InferRequest) (string, error) {
	res, err := p.client.CreateMessagesStream(ctx, anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: req.Config.ModelName,
			Messages: []anthropic.Message{
				{
					Role: anthropic.RoleUser,
					Content: []anthropic.MessageContent{
						anthropic.NewTextMessageContent(req.Message),
					},
				},
			},
			MaxTokens:   req.MessageOptions.MaxTokens,
			Temperature: &req.MessageOptions.Temperature,
		},
	})
	if err != nil {
		return "", errors.Wrap(err, "anthropic messages stream error")
	}

	slog.Debug("got response from anthropic", "model", res.Model, "res", res.GetFirstContentText(), "req", req.Message)
	return res.GetFirstContentText(), nil
}

func (p *Provider) GenerateResponseAsync(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	outChan := make(chan llm.StreamDelta)
	go func() {
		defer close(outChan)
		_, err := p.client.CreateMessagesStream(ctx, anthropic.MessagesStreamRequest{
			MessagesRequest: anthropic.MessagesRequest{
				Model: req.Config.ModelName,
				Messages: []anthropic.Message{
					{
						Role: anthropic.RoleUser,
						Content: []anthropic.MessageContent{
							anthropic.NewTextMessageContent(req.Message),
						},
					},
				},
				MaxTokens:   req.MessageOptions.MaxTokens,
				Temperature: &req.MessageOptions.Temperature,
			},
			OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
				if data.Delta.Text == "" {
					outChan <- llm.StreamDelta{
						EOF: true,
					}
					return
				}

				outChan <- llm.StreamDelta{
					Text: data.Delta.Text,
				}
			},
		})
		if err != nil {
			slog.Error("anthropic messages stream error", "err", err)
			return
		}
	}()

	return outChan, nil
}
