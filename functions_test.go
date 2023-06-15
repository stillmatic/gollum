package gollum_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"github.com/stillmatic/gollum"
	"github.com/stretchr/testify/assert"
)

type addInput struct {
	A int `json:"a" json_schema:"required"`
	B int `json:"b" json_schema:"required"`
}
type addOutput struct {
	C int `json:"c"`
}

// add_ is a unnecessary testing function that adds two numbers
func add_(input addInput) (addOutput, error) {
	return addOutput{C: input.A + input.B}, nil
}

type getWeatherInput struct {
	Location string `json:"location" jsonschema_description:"The city and state, e.g. San Francisco, CA" jsonschema:"required"`
	Unit     string `json:"unit,omitempty" jsonschema:"enum=celsius,enum=fahrenheit" jsonschema_description:"The unit of temperature"`
}

func TestConstructJSONSchema(t *testing.T) {
	t.Run("add_", func(t *testing.T) {
		res := gollum.StructToJsonSchema("add_", "adds two numbers", addInput{})
		expectedStr := `{"name":"add_","description":"adds two numbers","parameters":{"properties":{"a":{"type":"integer"},"b":{"type":"integer"}},"type":"object","required":["a","b"]}}`
		b, err := json.Marshal(res)
		assert.NoError(t, err)
		assert.Equal(t, expectedStr, string(b))
	})
	t.Run("getWeather", func(t *testing.T) {
		res := gollum.StructToJsonSchema("getWeather", "Get the current weather in a given location", getWeatherInput{})
		assert.Equal(t, res.Name, "getWeather")
		assert.Equal(t, res.Description, "Get the current weather in a given location")
		// assert.Equal(t, res.Parameters.Type, "object")
		expectedStr := `{"name":"getWeather","description":"Get the current weather in a given location","parameters":{"properties":{"location":{"type":"string","description":"The city and state, e.g. San Francisco, CA"},"unit":{"type":"string","enum":["celsius","fahrenheit"],"description":"The unit of temperature"}},"type":"object","required":["location"]}}`
		b, err := json.Marshal(res)
		assert.NoError(t, err)
		assert.Equal(t, expectedStr, string(b))
	})
}

type chatCompletionMessage struct {
	openai.ChatCompletionMessage
	FunctionCall functionCall `json:"function_call,omitempty"`
}

type functionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type chatCompletionRequest struct {
	// include the original fields
	openai.ChatCompletionRequest
	// Function stufff -- this is the part we care about
	Functions    []gollum.FunctionInput `json:"functions,omitempty"`
	FunctionCall string                 `json:"function_call,omitempty"`
}

type chatCompletionChoice struct {
	Index   int                   `json:"index"`
	Message chatCompletionMessage `json:"message"`
}

type chatCompletionResponse struct {
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

func (api *TestAPI) SendRequest(ctx context.Context, chatRequest chatCompletionRequest) (*chatCompletionResponse, error) {
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

	var chatResponse chatCompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResponse)
	if err != nil {
		return nil, err
	}

	return &chatResponse, nil
}

func TestEndToEnd(t *testing.T) {
	godotenv.Load()
	baseAPIURL := "https://api.openai.com/v1/chat/completions"
	openAIKey := os.Getenv("OPENAI_API_KEY")
	assert.NotEmpty(t, openAIKey)

	api := NewTestAPI(baseAPIURL, openAIKey)
	t.Run("weather", func(t *testing.T) {
		fi := gollum.StructToJsonSchema("weather", "Get the current weather in a given location", getWeatherInput{})

		chatRequest := chatCompletionRequest{
			ChatCompletionRequest: openai.ChatCompletionRequest{
				Model: "gpt-3.5-turbo-0613",
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    "user",
						Content: "Whats the temperature in Boston?",
					},
				},
				MaxTokens:   256,
				Temperature: 0.0,
			},
			Functions: []gollum.FunctionInput{fi},
		}

		ctx := context.Background()
		resp, err := api.SendRequest(ctx, chatRequest)
		assert.NoError(t, err)

		assert.Equal(t, resp.Model, "gpt-3.5-turbo-0613")
		assert.NotEmpty(t, resp.Choices)
		assert.Empty(t, resp.Choices[0].Message.Content)
		assert.NotNil(t, resp.Choices[0].Message.FunctionCall)
		assert.Equal(t, resp.Choices[0].Message.FunctionCall.Name, "weather")

		// this is somewhat flaky - about 20% of the time it returns 'Boston'
		expectedArg := `{"location": "Boston, MA"}`
		parser := gollum.NewJSONParser[getWeatherInput](false)
		expectedStruct, err := parser.Parse(ctx, expectedArg)
		assert.NoError(t, err)
		input, err := parser.Parse(ctx, resp.Choices[0].Message.FunctionCall.Arguments)
		assert.NoError(t, err)
		assert.Equal(t, expectedStruct, input)
	})

	t.Run("callOpenAI", func(t *testing.T) {
		fi := gollum.StructToJsonSchema("ChatCompletion", "Call the OpenAI chat completion API", chatCompletionRequest{})

		chatRequest := chatCompletionRequest{
			ChatCompletionRequest: openai.ChatCompletionRequest{
				Model: "gpt-3.5-turbo-0613",
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: "Construct a ChatCompletionRequest to answer the user's question, but using Kirby references. Do not answer the question directly using prior knowledge, you must generate a ChatCompletionRequest that will answer the question.",
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "What is the definition of recursion?",
					},
				},
				MaxTokens:   256,
				Temperature: 0.0,
			},
			Functions: []gollum.FunctionInput{fi},
		}

		ctx := context.Background()
		resp, err := api.SendRequest(ctx, chatRequest)
		assert.NoError(t, err)
		assert.Equal(t, resp.Model, "gpt-3.5-turbo-0613")
		assert.NotEmpty(t, resp.Choices)
		assert.Empty(t, resp.Choices[0].Message.Content)
		assert.NotNil(t, resp.Choices[0].Message.FunctionCall)
		assert.Equal(t, resp.Choices[0].Message.FunctionCall.Name, "ChatCompletion")

		parser := gollum.NewJSONParser[openai.ChatCompletionRequest](false)
		input, err := parser.Parse(ctx, resp.Choices[0].Message.FunctionCall.Arguments)
		assert.NoError(t, err)
		assert.NotEmpty(t, input)

		// an example output:
		//  "{
		//   "model": "gpt-3.5-turbo",
		//   "messages": [
		//     {"role": "system", "content": "You are Kirby, a friendly virtual assistant."},
		//     {"role": "user", "content": "What is the definition of recursion?"}
		//   ]
		// }"
	})
}
