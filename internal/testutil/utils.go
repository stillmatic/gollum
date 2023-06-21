package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sashabaranov/go-openai"
	"github.com/stillmatic/gollum"
)

type ChatCompletionMessage struct {
	openai.ChatCompletionMessage
	FunctionCall FunctionCall `json:"function_call,omitempty"`
}

type FunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type ChatCompletionRequest struct {
	// include the original fields
	openai.ChatCompletionRequest
	// Function stufff -- this is the part we care about
	Functions    []gollum.FunctionInput `json:"functions,omitempty"`
	FunctionCall string                 `json:"function_call,omitempty"`
}

type chatCompletionChoice struct {
	Index   int                   `json:"index"`
	Message ChatCompletionMessage `json:"message"`
}

type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []chatCompletionChoice `json:"choices"`
}

type TestAPI struct {
	baseAPIURL string
	apiKey     string
	client     *http.Client
}

func NewTestAPI(baseAPIURL, apiKey string) *TestAPI {
	return &TestAPI{
		baseAPIURL: baseAPIURL,
		apiKey:     apiKey,
		client:     &http.Client{},
	}
}

func (api *TestAPI) SendRequest(ctx context.Context, chatRequest ChatCompletionRequest) (*ChatCompletionResponse, error) {
	b, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", api.baseAPIURL, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+api.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var chatResponse ChatCompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResponse)
	if err != nil {
		return nil, err
	}

	return &chatResponse, nil
}
