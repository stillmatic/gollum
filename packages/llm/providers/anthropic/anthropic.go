package anthropic

import (
	"context"
	"encoding/base64"
	"github.com/stillmatic/gollum/packages/llm"
	"log/slog"
	"slices"

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

func reqToMessages(req llm.InferRequest) ([]anthropic.Message, error) {
	msgs := make([]anthropic.Message, 0)
	for _, m := range req.Messages {
		// only allow user and assistant roles
		// TODO: this should be a little cleaner...
		if !(slices.Index([]string{anthropic.RoleUser, anthropic.RoleAssistant}, m.Role) > -1) {
			return nil, errors.New("invalid role")
		}
		content := make([]anthropic.MessageContent, 0)
		content = append(content, anthropic.NewTextMessageContent(m.Content))
		if m.Image != nil && len(*m.Image) > 0 {
			b64Image := base64.StdEncoding.EncodeToString(*m.Image)
			// TODO: support other image types
			content = append(content, anthropic.NewImageMessageContent(
				anthropic.MessageContentImageSource{Type: "base64", MediaType: "image/png", Data: b64Image}))
		}
		msgs = append(msgs, anthropic.Message{
			Role: m.Role,
			Content: []anthropic.MessageContent{
				anthropic.NewTextMessageContent(m.Content),
			},
		})
	}

	return msgs, nil
}

func (p *Provider) GenerateResponse(ctx context.Context, req llm.InferRequest) (string, error) {
	msgs, err := reqToMessages(req)
	if err != nil {
		return "", errors.Wrap(err, "invalid messages")
	}
	res, err := p.client.CreateMessagesStream(ctx, anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model:       req.ModelConfig.ModelName,
			Messages:    msgs,
			MaxTokens:   req.MessageOptions.MaxTokens,
			Temperature: &req.MessageOptions.Temperature,
		},
	})
	if err != nil {
		return "", errors.Wrap(err, "anthropic messages stream error")
	}

	return res.GetFirstContentText(), nil
}

func (p *Provider) GenerateResponseAsync(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	outChan := make(chan llm.StreamDelta)
	go func() {
		defer close(outChan)
		msgs, err := reqToMessages(req)
		if err != nil {
			slog.Error("invalid messages", "err", err)
			return
		}

		_, err = p.client.CreateMessagesStream(ctx, anthropic.MessagesStreamRequest{
			MessagesRequest: anthropic.MessagesRequest{
				Model:       req.ModelConfig.ModelName,
				Messages:    msgs,
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
