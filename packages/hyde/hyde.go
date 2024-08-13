package hyde

import (
	"context"
	"fmt"
	"github.com/stillmatic/gollum/packages/vectorstore"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/stillmatic/gollum"
	"github.com/viterin/vek/vek32"
)

type Prompter interface {
	BuildPrompt(context.Context, string) string
}

type Generator interface {
	Generate(ctx context.Context, input string, n int) ([]string, error)
}

type Encoder interface {
	Encode(context.Context, string) ([]float32, error)
	EncodeBatch(context.Context, []string) ([][]float32, error)
}

type Searcher interface {
	Search(context.Context, []float32, int) ([]*gollum.Document, error)
}

type Hyde struct {
	prompter  Prompter
	generator Generator
	encoder   Encoder
	searcher  Searcher
}

type ZeroShotPrompter struct {
	template string
}

func NewZeroShotPrompter(template string) *ZeroShotPrompter {
	// something like
	// Roleplay as a character. Write a short biographical answer to the question.
	// Q: %s
	// A:
	return &ZeroShotPrompter{
		template: template,
	}
}

func (z *ZeroShotPrompter) BuildPrompt(ctx context.Context, prompt string) string {
	// fill in template values
	return fmt.Sprintf(z.template, prompt)
}

type LLMGenerator struct {
	Model gollum.ChatCompleter
}

func NewLLMGenerator(model gollum.ChatCompleter) *LLMGenerator {
	return &LLMGenerator{
		Model: model,
	}
}

func (l *LLMGenerator) Generate(ctx context.Context, prompt string, n int) ([]string, error) {
	createChatCompletionReq := openai.ChatCompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Model: openai.GPT3Dot5Turbo,
		N:     n,
		// hyperparams from https://github.com/texttron/hyde/blob/74101c5157e04f7b57559e7da8ef4a4e5b6da82b/src/hyde/generator.py#LL15C121-L15C121
		Temperature:      0.9,
		MaxTokens:        512,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		Stop:             []string{"\n\n\n"},
	}
	resp, err := l.Model.CreateChatCompletion(ctx, createChatCompletionReq)
	if err != nil {
		return make([]string, 0), err
	}
	replies := make([]string, n)
	for i, choice := range resp.Choices {
		replies[i] = choice.Message.Content
	}
	return replies, nil
}

type LLMEncoder struct {
	Model gollum.Embedder
}

func NewLLMEncoder(model gollum.Embedder) *LLMEncoder {
	return &LLMEncoder{
		Model: model,
	}
}

func (l *LLMEncoder) Encode(ctx context.Context, query string) ([]float32, error) {
	createEmbeddingReq := openai.EmbeddingRequest{
		Input: []string{query},
		// TODO: allow customization
		Model: openai.AdaEmbeddingV2,
	}
	resp, err := l.Model.CreateEmbeddings(ctx, createEmbeddingReq)
	if err != nil {
		return make([]float32, 0), err
	}
	return resp.Data[0].Embedding, nil
}

func (l *LLMEncoder) EncodeBatch(ctx context.Context, docs []string) ([][]float32, error) {
	createEmbeddingReq := openai.EmbeddingRequest{
		Input: docs,
		// TODO: allow customization
		Model: openai.AdaEmbeddingV2,
	}
	resp, err := l.Model.CreateEmbeddings(ctx, createEmbeddingReq)
	if err != nil {
		return make([][]float32, 0), err
	}
	embeddings := make([][]float32, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}
	return embeddings, nil
}

type VectorSearcher struct {
	vs vectorstore.VectorStore
}

func NewVectorSearcher(vs vectorstore.VectorStore) *VectorSearcher {
	return &VectorSearcher{
		vs: vs,
	}
}

func (v *VectorSearcher) Search(ctx context.Context, query []float32, n int) ([]*gollum.Document, error) {
	qb := vectorstore.QueryRequest{
		EmbeddingFloats: query,
		K:               n,
	}
	return v.vs.Query(ctx, qb)
}

func NewHyde(prompter Prompter, generator Generator, encoder Encoder, searcher Searcher) *Hyde {
	return &Hyde{
		prompter:  prompter,
		generator: generator,
		encoder:   encoder,
		searcher:  searcher,
	}
}

// Prompt builds a prompt from the given string.
func (h *Hyde) Prompt(ctx context.Context, prompt string) string {
	return h.prompter.BuildPrompt(ctx, prompt)
}

// Generate generates n hypothesis documents.
func (h *Hyde) Generate(ctx context.Context, prompt string, n int) ([]string, error) {
	searchPrompt := h.prompter.BuildPrompt(ctx, prompt)
	return h.generator.Generate(ctx, searchPrompt, n)
}

// Encode encodes a query and a list of documents into a single embedding.
func (h *Hyde) Encode(ctx context.Context, query string, docs []string) ([]float32, error) {
	docs = append(docs, query)
	embeddings, err := h.encoder.EncodeBatch(ctx, docs)
	if err != nil {
		return make([]float32, 0), errors.Wrap(err, "error encoding batch")
	}
	embedDim := len(embeddings[0])
	numEmbeddings := float32(len(embeddings))
	// mean pooling of response embeddings
	avgEmbedding := make([]float32, embedDim)
	for _, embedding := range embeddings {
		vek32.Add_Inplace(avgEmbedding, embedding)
	}
	// unclear if this is faster or slower than naive approach. a little cleaner code though.
	vek32.DivNumber_Inplace(avgEmbedding, numEmbeddings)
	return avgEmbedding, nil
}

func (h *Hyde) Search(ctx context.Context, hydeVector []float32, k int) ([]*gollum.Document, error) {
	return h.searcher.Search(ctx, hydeVector, k)
}

func (h *Hyde) SearchEndToEnd(ctx context.Context, query string, k int) ([]*gollum.Document, error) {
	// generate n hypothesis documents
	docs, err := h.Generate(ctx, query, k)
	if err != nil {
		return nil, errors.Wrap(err, "error generating hypothetical documents")
	}
	// encode query and hypothesis documents
	hydeVector, err := h.Encode(ctx, query, docs)
	if err != nil {
		return make([]*gollum.Document, 0), errors.Wrap(err, "error encoding hypothetical documents")
	}
	// search for the most similar documents
	return h.Search(ctx, hydeVector, k)
}
