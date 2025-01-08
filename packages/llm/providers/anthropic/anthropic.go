package anthropic

import (
	"context"
	"encoding/base64"
	"log/slog"
	"slices"

	"github.com/stillmatic/gollum/packages/llm"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/pkg/errors"
)

type Provider struct {
	client       *anthropic.Client
	cacheEnabled bool
}

func NewAnthropicProvider(apiKey string) *Provider {
	return &Provider{
		client: anthropic.NewClient(apiKey),
	}
}

func NewAnthropicProviderWithCache(apiKey string) *Provider {
	client := anthropic.NewClient(apiKey, anthropic.WithBetaVersion(anthropic.BetaPromptCaching20240731))
	return &Provider{
		client:       client,
		cacheEnabled: true,
	}
}

func reqToMessages(req llm.InferRequest) ([]anthropic.Message, []anthropic.MessageSystemPart, error) {
	msgs := make([]anthropic.Message, 0)
	systemMsgs := make([]anthropic.MessageSystemPart, 0)
	for _, m := range req.Messages {
		if m.Role == "system" {
			msgContent := anthropic.MessageSystemPart{
				Type: "text",
				Text: m.Content,
			}
			if m.ShouldCache {
				msgContent.CacheControl = &anthropic.MessageCacheControl{
					Type: anthropic.CacheControlTypeEphemeral,
				}
			}
			systemMsgs = append(systemMsgs, msgContent)
			continue
		}

		// only allow user and assistant roles
		// TODO: this should be a little cleaner...
		if !(slices.Index([]string{string(anthropic.RoleUser), string(anthropic.RoleAssistant)}, m.Role) > -1) {
			return nil, nil, errors.New("invalid role")
		}
		content := make([]anthropic.MessageContent, 0)
		txtContent := anthropic.NewTextMessageContent(m.Content)
		// this will fail if the model is not configured to cache
		if m.ShouldCache {
			txtContent.SetCacheControl()
		}
		content = append(content, txtContent)
		if m.Image != nil && len(m.Image) > 0 {
			b64Image := base64.StdEncoding.EncodeToString(m.Image)
			// TODO: support other image types
			content = append(content, anthropic.NewImageMessageContent(
				anthropic.MessageContentSource{Type: "base64", MediaType: "image/png", Data: b64Image}))
		}
		newMsg := anthropic.Message{
			Role:    anthropic.ChatRole(m.Role),
			Content: content,
		}

		msgs = append(msgs, newMsg)
	}

	return msgs, systemMsgs, nil
}

func (p *Provider) GenerateResponse(ctx context.Context, req llm.InferRequest) (string, error) {
	msgs, systemPrompt, err := reqToMessages(req)
	if err != nil {
		return "", errors.Wrap(err, "invalid messages")
	}
	msgsReq := anthropic.MessagesRequest{
		Model:       anthropic.Model(req.ModelConfig.ModelName),
		Messages:    msgs,
		MaxTokens:   req.MessageOptions.MaxTokens,
		Temperature: &req.MessageOptions.Temperature,
	}
	if systemPrompt != nil && len(systemPrompt) > 0 {
		msgsReq.MultiSystem = systemPrompt
	}
	res, err := p.client.CreateMessagesStream(ctx, anthropic.MessagesStreamRequest{
		MessagesRequest: msgsReq,
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
		msgs, systemPrompt, err := reqToMessages(req)
		if err != nil {
			slog.Error("invalid messages", "err", err)
			return
		}
		msgsReq := anthropic.MessagesRequest{
			Model:       anthropic.Model(req.ModelConfig.ModelName),
			Messages:    msgs,
			MaxTokens:   req.MessageOptions.MaxTokens,
			Temperature: &req.MessageOptions.Temperature,
		}
		if systemPrompt != nil {
			msgsReq.MultiSystem = systemPrompt
		}

		_, err = p.client.CreateMessagesStream(ctx, anthropic.MessagesStreamRequest{
			MessagesRequest: msgsReq,
			OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
				if data.Delta.Text == nil {
					outChan <- llm.StreamDelta{
						EOF: true,
					}
					return
				}

				outChan <- llm.StreamDelta{
					Text: *data.Delta.Text,
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
