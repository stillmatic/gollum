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

	// it is slightly better to build a trie, indexed on hashes of each message
	// since we can quickly get based on the prefix (i.e. existing messages)
	// but ... your number of messages is probably not THAT high to justify the complexity.
	messagesToCache := make([]llm.InferMessage, 0)
	for _, message := range req.Messages {
		if message.ShouldCache {
			messagesToCache = append(messagesToCache, message)
		}
	}
	model := p.getModel(req)
	if len(messagesToCache) > 0 {
		// hash the messages and check if the overall object is cached.
		// we choose to do this because you may have a later message identical to an earlier message
		// if we find exact match for this set of messages, load it.
		hashKeys := make([]string, 0, len(messagesToCache))
		for _, message := range messagesToCache {
			// it is possible to have collision between user + assistant content being identical
			// this feels like a rare case especially given that we are ordering sensitive in the hash.
			hashKeys = append(hashKeys, getHash(message.Content))
		}
		joinedKey := strings.Join(hashKeys, "/")
		var cachedContent *genai.CachedContent
		// if the cached object exists, load it
		if _, ok := p.cachedContentMap[joinedKey]; ok {
			cachedContent, _ = p.client.GetCachedContent(ctx, joinedKey)
			model = p.client.GenerativeModelFromCachedContent(cachedContent)
		} else {
			// otherwise, create a new cached object
			cc, err := p.createCachedContent(ctx, joinedKey, req.ModelConfig.ModelName)
			if err != nil {
				return "", errors.Wrap(err, "google upload file error")
			}
			model = p.client.GenerativeModelFromCachedContent(cc)
		}
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

func (p *Provider) generateResponseChat(ctx context.Context, req llm.InferRequest) (string, error) {
	model := p.getModel(req)

	// annoyingly, the last message is the one we want to generate a response to, so we need to split it out
	msgs, sysInstr := multiTurnMessageToParts(req.Messages[:len(req.Messages)-1])
	if sysInstr != nil {
		model.SystemInstruction = sysInstr
	}

	cs := model.StartChat()
	cs.History = msgs
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
		msgs, sysInstr := multiTurnMessageToParts(req.Messages[:len(req.Messages)-1])
		if sysInstr != nil {
			model.SystemInstruction = sysInstr
		}
		cs := model.StartChat()
		cs.History = msgs

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

// GenerateEmbedding generates embeddings for the given input.
//
// NB chua: This is a confusing method in the docs.
// - There are two separate API methods and it's unclear which you should use. Is batch with 1 the same as single?
// - What's the maximum number of docs to embed at once?
// - TaskType is automatically set..? I don't see how to configure it ...?
// see also https://pkg.go.dev/github.com/google/generative-ai-go/genai#TaskType
func (p *Provider) GenerateEmbedding(ctx context.Context, req llm.EmbedRequest) (*llm.EmbeddingResponse, error) {
	em := p.client.EmbeddingModel(req.ModelConfig.ModelName)

	// if there is only one input, use the single API
	if len(req.Input) == 1 {
		resp, err := em.EmbedContent(ctx, genai.Text(req.Input[0]))
		if err != nil {
			return nil, errors.Wrap(err, "google embedding error")
		}

		return &llm.EmbeddingResponse{Data: []llm.Embedding{{Values: resp.Embedding.Values}}}, nil
	}
	// otherwise, use the batch API. I'm not sure there's much difference though...
	batchReq := em.NewBatch()
	for _, input := range req.Input {
		batchReq.AddContent(genai.Text(input))
	}
	resp, err := em.BatchEmbedContents(ctx, batchReq)
	if err != nil {
		return nil, errors.Wrap(err, "google batch embedding error")
	}

	respVectors := make([]llm.Embedding, len(resp.Embeddings))
	for i, v := range resp.Embeddings {
		respVectors[i] = llm.Embedding{
			Values: v.Values,
		}
	}

	return &llm.EmbeddingResponse{
		Data: respVectors,
	}, nil
}
