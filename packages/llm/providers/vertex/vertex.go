// Package vertex implements the Vertex AI api
// it is largely similar to the Google "ai studio" provider but uses a different library...
package vertex

import (
	"cloud.google.com/go/vertexai/genai"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stillmatic/gollum/packages/llm"
	"google.golang.org/api/iterator"
	"log"
)

type VertexAIProvider struct {
	client *genai.Client
}

func NewVertexAIProvider(ctx context.Context, projectID, location string) (*VertexAIProvider, error) {
	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Vertex AI client")
	}

	ccIter := client.ListCachedContents(ctx)
	for {
		cc, err := ccIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to list cached contents")
		}
		log.Printf("Cached content: %v", cc)
	}

	return &VertexAIProvider{
		client: client,
	}, nil
}

func (p *VertexAIProvider) getModel(req llm.InferRequest) *genai.GenerativeModel {
	// this does NOT validate if the model name is valid, that is done at inference time.
	model := p.client.GenerativeModel(req.ModelConfig.ModelName)
	model.SetTemperature(req.MessageOptions.Temperature)
	model.SetMaxOutputTokens(int32(req.MessageOptions.MaxTokens))

	return model
}

func (p *VertexAIProvider) GenerateResponse(ctx context.Context, req llm.InferRequest) (string, error) {
	if len(req.Messages) > 1 {
		return p.generateResponseMultiTurn(ctx, req)
	}
	return p.generateResponseSingleTurn(ctx, req)
}

func (p *VertexAIProvider) generateResponseSingleTurn(ctx context.Context, req llm.InferRequest) (string, error) {
	model := p.getModel(req)
	parts := messageToParts(req.Messages[0])

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate content")
	}

	return flattenResponse(resp), nil
}

func (p *VertexAIProvider) generateResponseMultiTurn(ctx context.Context, req llm.InferRequest) (string, error) {
	model := p.getModel(req)

	msgs, sysInstr := multiTurnMessageToParts(req.Messages[:len(req.Messages)-1])
	if sysInstr != nil {
		model.SystemInstruction = sysInstr
	}

	cs := model.StartChat()
	cs.History = msgs
	mostRecentMessage := req.Messages[len(req.Messages)-1]

	// Send the last message
	resp, err := cs.SendMessage(ctx, genai.Text(mostRecentMessage.Content))
	if err != nil {
		return "", errors.Wrap(err, "failed to send message in chat")
	}

	return flattenResponse(resp), nil
}

func (p *VertexAIProvider) GenerateResponseAsync(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	if len(req.Messages) > 1 {
		return p.generateResponseAsyncMultiTurn(ctx, req)
	}
	return p.generateResponseAsyncSingleTurn(ctx, req)
}

func (p *VertexAIProvider) generateResponseAsyncSingleTurn(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	outChan := make(chan llm.StreamDelta)

	go func() {
		defer close(outChan)

		model := p.getModel(req)
		parts := messageToParts(req.Messages[0])

		iter := model.GenerateContentStream(ctx, parts...)

		for {
			resp, err := iter.Next()
			if errors.Is(err, iterator.Done) {
				outChan <- llm.StreamDelta{EOF: true}
				break
			}
			if err != nil {
				log.Printf("Error from Vertex AI stream: %v", err)
				return
			}

			content := flattenResponse(resp)
			if content != "" {
				select {
				case <-ctx.Done():
					return
				case outChan <- llm.StreamDelta{Text: content}:
				}
			}
		}
	}()

	return outChan, nil
}

func (p *VertexAIProvider) generateResponseAsyncMultiTurn(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	outChan := make(chan llm.StreamDelta)

	go func() {
		defer close(outChan)

		model := p.getModel(req)
		cs := model.StartChat()

		// Add previous messages to chat history
		for _, msg := range req.Messages[:len(req.Messages)-1] {
			parts := messageToParts(msg)
			cs.History = append(cs.History, &genai.Content{
				Parts: parts,
				Role:  msg.Role,
			})
		}

		// Send the last message
		lastMsg := req.Messages[len(req.Messages)-1]
		iter := cs.SendMessageStream(ctx, messageToParts(lastMsg)...)

		for {
			resp, err := iter.Next()
			if errors.Is(err, iterator.Done) {
				outChan <- llm.StreamDelta{EOF: true}
				break
			}
			if err != nil {
				log.Printf("Error from Vertex AI stream: %v", err)
				return
			}

			content := flattenResponse(resp)
			if content != "" {
				select {
				case <-ctx.Done():
					return
				case outChan <- llm.StreamDelta{Text: content}:
				}
			}
		}
	}()

	return outChan, nil
}

func messageToParts(message llm.InferMessage) []genai.Part {
	parts := []genai.Part{genai.Text(message.Content)}
	if message.Image != nil && len(message.Image) > 0 {
		parts = append(parts, genai.ImageData("png", message.Image))
	}
	return parts
}

func multiTurnMessageToParts(messages []llm.InferMessage) ([]*genai.Content, *genai.Content) {
	sysInstructionParts := make([]genai.Part, 0)
	hist := make([]*genai.Content, 0, len(messages))
	for _, message := range messages {
		parts := []genai.Part{genai.Text(message.Content)}
		if message.Image != nil && len(message.Image) > 0 {
			parts = append(parts, genai.ImageData("png", message.Image))
		}
		if message.Role == "system" {
			sysInstructionParts = append(sysInstructionParts, parts...)
			continue
		}
		hist = append(hist, &genai.Content{
			Parts: parts,
			Role:  message.Role,
		})
	}
	if len(sysInstructionParts) > 0 {
		return hist, &genai.Content{
			Parts: sysInstructionParts,
		}
	}

	return hist, nil
}

func flattenResponse(resp *genai.GenerateContentResponse) string {
	var result string
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			result += fmt.Sprintf("%v", part)
		}
	}
	return result
}

var _ llm.Responder = &VertexAIProvider{}
