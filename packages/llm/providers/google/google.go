package google

import (
	"context"
	"github.com/cespare/xxhash/v2"
	"github.com/stillmatic/gollum/packages/llm"
	"google.golang.org/api/option"
	"log/slog"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

type Provider struct {
	client           *genai.Client
	cachedFileMap    map[string]string
	cachedContentMap map[string]struct{}
}

func NewGoogleProvider(ctx context.Context, apiKey string) (*Provider, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, errors.Wrap(err, "google client error")
	}

	// load cached content map
	p := &Provider{client: client}
	err = p.refreshCachedContentMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "google refresh cached content map error")
	}

	return p, nil
}

func (p *Provider) refreshCachedFileMap(ctx context.Context) error {
	iter := p.client.ListFiles(ctx)
	cachedFileMap := make(map[string]string)
	for {
		cachedFile, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return errors.Wrap(err, "google list cached files error")
		}
		cachedFileMap[cachedFile.Name] = cachedFile.URI
	}
	p.cachedFileMap = cachedFileMap
	return nil
}

func (p *Provider) refreshCachedContentMap(ctx context.Context) error {
	iter := p.client.ListCachedContents(ctx)
	cachedContentMap := make(map[string]struct{})
	for {
		cachedContent, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return errors.Wrap(err, "google list cached content error")
		}
		cachedContentMap[cachedContent.Name] = struct{}{}
	}
	p.cachedContentMap = cachedContentMap
	return nil
}

func getHash(value string) string {
	return string(xxhash.New().Sum([]byte(value)))
}

func (p *Provider) uploadFile(ctx context.Context, key string, value string) (*genai.File, error) {
	// check if the file is already cached
	if _, ok := p.cachedFileMap[key]; ok {
		// if so, load the cached file and return it
		cachedFile, err := p.client.GetFile(ctx, key)
		if err != nil {
			return nil, errors.Wrap(err, "google get file error")
		}
		return cachedFile, nil
	}

	r := strings.NewReader(value)
	file, err := p.client.UploadFile(ctx, key, r, nil)
	if err != nil {
		return nil, errors.Wrap(err, "google upload file error")
	}
	p.cachedFileMap[key] = file.URI

	return file, nil
}

func (p *Provider) createCachedContent(ctx context.Context, value string, modelName string) (*genai.CachedContent, error) {
	key := getHash(value)

	// check if the content is already cached
	if _, ok := p.cachedContentMap[key]; ok {
		// if so, load the cached content and return it
		cachedContent, err := p.client.GetCachedContent(ctx, key)
		if err != nil {
			return nil, errors.Wrap(err, "google get cached content error")
		}
		return cachedContent, nil
	}

	file, err := p.uploadFile(ctx, key, value)
	if err != nil {
		return nil, errors.Wrap(err, "error uploading file")
	}
	fd := genai.FileData{URI: file.URI}
	cc := &genai.CachedContent{
		Name:     key,
		Model:    modelName,
		Contents: []*genai.Content{genai.NewUserContent(fd)},
		// TODO: make this configurable
		// maybe something like an optional field, 'ephemeral' / 'hour'?
		// default matches Anthropic's 5 minute TTL
		Expiration: genai.ExpireTimeOrTTL{TTL: 5 * time.Minute},
	}
	content, err := p.client.CreateCachedContent(ctx, cc)
	if err != nil {
		return nil, errors.Wrap(err, "error creating cached content")
	}

	return content, nil
}

func (p *Provider) getModel(req llm.InferRequest) *genai.GenerativeModel {
	model := p.client.GenerativeModel(req.ModelConfig.ModelName)
	model.SetTemperature(req.MessageOptions.Temperature)
	model.SetMaxOutputTokens(int32(req.MessageOptions.MaxTokens))
	// lol...
	model.SafetySettings = []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockNone},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockNone},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockNone},
		{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockNone},
	}
	model.SetCandidateCount(1)
	return model
}

func (p *Provider) GenerateResponse(ctx context.Context, req llm.InferRequest) (string, error) {
	if len(req.Messages) > 1 {
		return p.generateResponseChat(ctx, req)
	}
	anyCache := false

	// TODO: should we handle multiple cached messages?
	// This is helpful for chat conversations, but isn't chat cringe?
	// It's trivial to cache multiple messages in the same CachedContent object
	// but how do we do retrieval of the cached object?
	// I guess we could build a trie, indexed on hashes of each message?
	var messageToCache llm.InferMessage
	for _, message := range req.Messages {
		if message.ShouldCache {
			anyCache = true
			messageToCache = message
		}
	}
	model := p.getModel(req)
	if anyCache {
		// cache the messages
		cc, err := p.createCachedContent(ctx, messageToCache.Content, req.ModelConfig.ModelName)
		if err != nil {
			return "", errors.Wrap(err, "google upload file error")
		}
		model = p.client.GenerativeModelFromCachedContent(cc)
	}

	parts := singleTurnMessageToParts(req.Messages[0])

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", errors.Wrap(err, "google generate content error")
	}
	respStr := flattenResponse(resp)

	return respStr, nil
}

func singleTurnMessageToParts(message llm.InferMessage) []genai.Part {
	parts := []genai.Part{genai.Text(message.Content)}
	if message.Image != nil && len(*message.Image) > 0 {
		parts = append(parts, genai.ImageData("png", *message.Image))
	}
	return parts
}

func multiTurnMessageToParts(messages []llm.InferMessage) []*genai.Content {
	hist := make([]*genai.Content, 0, len(messages))
	for _, message := range messages {
		parts := []genai.Part{genai.Text(message.Content)}
		if message.Image != nil && len(*message.Image) > 0 {
			parts = append(parts, genai.ImageData("png", *message.Image))
		}
		hist = append(hist, &genai.Content{
			Parts: parts,
			Role:  message.Role,
		})
	}
	return hist
}

func (p *Provider) generateResponseChat(ctx context.Context, req llm.InferRequest) (string, error) {
	model := p.getModel(req)
	cs := model.StartChat()

	// annoyingly, the last message is the one we want to generate a response to, so we need to split it out
	cs.History = multiTurnMessageToParts(req.Messages[:len(req.Messages)-1])
	mostRecentMessage := req.Messages[len(req.Messages)-1]

	// NB chua: this might be a bug but Google doesn't seem to accept multiple parts in the same message
	// in the chat API. So can't send text + image if it exists.
	//mostRecentMessagePart := []genai.Part{genai.Text(mostRecentMessage.Content)}
	//if mostRecentMessage.Image != nil && len(*mostRecentMessage.Image) > 0 {
	//	mostRecentMessagePart = append(mostRecentMessagePart, genai.ImageData("png", *mostRecentMessage.Image))
	//}

	resp, err := cs.SendMessage(ctx, genai.Text(mostRecentMessage.Content))
	if err != nil {
		return "", errors.Wrap(err, "google generate content error")
	}
	respStr := flattenResponse(resp)

	return respStr, nil
}

func (p *Provider) GenerateResponseAsync(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	if len(req.Messages) > 1 {
		return p.generateResponseAsyncChat(ctx, req)
	}
	return p.generateResponseAsyncSingle(ctx, req)
}

func (p *Provider) generateResponseAsyncSingle(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	outChan := make(chan llm.StreamDelta)

	go func() {
		defer close(outChan)

		model := p.getModel(req)

		parts := singleTurnMessageToParts(req.Messages[0])
		iter := model.GenerateContentStream(ctx, parts...)

		for {
			resp, err := iter.Next()
			if errors.Is(err, iterator.Done) {
				outChan <- llm.StreamDelta{EOF: true}
				break
			}
			if err != nil {
				slog.Error("error from gemini stream", "err", err, "req", req.Messages[0].Content, "model", req.ModelConfig.ModelName)
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

func (p *Provider) generateResponseAsyncChat(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	outChan := make(chan llm.StreamDelta)

	go func() {
		defer close(outChan)

		model := p.getModel(req)
		cs := model.StartChat()

		cs.History = multiTurnMessageToParts(req.Messages[:len(req.Messages)-1])
		mostRecentMessage := req.Messages[len(req.Messages)-1]

		iter := cs.SendMessageStream(ctx, genai.Text(mostRecentMessage.Content))

		for {
			resp, err := iter.Next()
			if errors.Is(err, iterator.Done) {
				outChan <- llm.StreamDelta{EOF: true}
				break
			}
			if err != nil {
				slog.Error("error from gemini stream", "err", err, "req", mostRecentMessage.Content, "model", req.ModelConfig.ModelName)
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

// flattenResponse flattens the response from the Gemini API into a single string.
func flattenResponse(resp *genai.GenerateContentResponse) string {
	var rtn strings.Builder
	for i, part := range resp.Candidates[0].Content.Parts {
		switch part := part.(type) {
		case genai.Text:
			if i > 0 {
				rtn.WriteString(" ")
			}
			rtn.WriteString(string(part))
		}
	}
	return rtn.String()
}
