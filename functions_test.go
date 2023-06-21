package gollum_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"github.com/stillmatic/gollum"
	"github.com/stillmatic/gollum/internal/testutil"
	"github.com/stretchr/testify/assert"
)

type addInput struct {
	A int `json:"a" json_schema:"required"`
	B int `json:"b" json_schema:"required"`
}

type getWeatherInput struct {
	Location string `json:"location" jsonschema_description:"The city and state." jsonschema:"required,example=San Francisco, CA"`
	Unit     string `json:"unit,omitempty" jsonschema:"enum=celsius,enum=fahrenheit" jsonschema_description:"The unit of temperature,default=fahrenheit"`
}

type counter struct {
	Count int      `json:"count" jsonschema:"required" jsonschema_description:"total number of words in sentence"`
	Words []string `json:"words" jsonschema:"required" jsonschema_description:"list of words in sentence"`
}

type blobNode struct {
	Name     string     `json:"name" jsonschema:"required"`
	Children []blobNode `json:"children,omitempty" jsonschema_description:"list of child nodes - only applicable if this is a directory"`
	NodeType string     `json:"node_type" jsonschema:"required,enum=file,enum=folder" jsonschema_description:"type of node, inferred from name"`
}

type queryNode struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"question to ask - questions can use information from children questions"`
	// NodeType string      `json:"node_type" jsonschema:"required,enum=single_question,enum=merge_responses" jsonschema_description:"type of question. Either a single question or a multi question merge when there are multiple questions."`
	Children []queryNode `json:"children,omitempty" jsonschema_description:"list of child questions that need to be answered before this question can be answered. Use a subquery when anything may be unknown, and we need to ask multiple questions to get the answer. Dependences must only be other queries."`
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

func TestEndToEnd(t *testing.T) {
	godotenv.Load()
	baseAPIURL := "https://api.openai.com/v1/chat/completions"
	openAIKey := os.Getenv("OPENAI_API_KEY")
	assert.NotEmpty(t, openAIKey)

	api := testutil.NewTestAPI(baseAPIURL, openAIKey)
	t.Run("weather", func(t *testing.T) {
		t.Skip("somewhat flaky - word counter is more reliable")
		fi := gollum.StructToJsonSchema("weather", "Get the current weather in a given location", getWeatherInput{})

		chatRequest := openai.ChatCompletionRequest{
			Model: "gpt-3.5-turbo-0613",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "user",
					Content: "Whats the temperature in Boston?",
				},
			},
			MaxTokens:   256,
			Temperature: 0.0,
			Functions:   []openai.FunctionDefinition{openai.FunctionDefinition(fi)},
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
		expectedArg := []byte(`{"location": "Boston, MA"}`)
		parser := gollum.NewJSONParserGeneric[getWeatherInput](false)
		expectedStruct, err := parser.Parse(ctx, expectedArg)
		assert.NoError(t, err)
		input, err := parser.Parse(ctx, []byte(resp.Choices[0].Message.FunctionCall.Arguments))
		assert.NoError(t, err)
		assert.Equal(t, expectedStruct, input)
	})

	t.Run("counter", func(t *testing.T) {
		fi := gollum.StructToJsonSchema("split_word", "Break sentences into words", counter{})
		chatRequest := openai.ChatCompletionRequest{
			Model: "gpt-3.5-turbo-0613",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "user",
					Content: "「What is the weather like in Boston?」Break down the above sentence into words",
				},
			},
			MaxTokens:   256,
			Temperature: 0.0,
			Functions: []openai.FunctionDefinition{
				openai.FunctionDefinition(fi),
			},
		}
		ctx := context.Background()
		resp, err := api.SendRequest(ctx, chatRequest)
		assert.NoError(t, err)

		assert.Equal(t, resp.Model, "gpt-3.5-turbo-0613")
		assert.NotEmpty(t, resp.Choices)
		assert.Empty(t, resp.Choices[0].Message.Content)
		assert.NotNil(t, resp.Choices[0].Message.FunctionCall)
		assert.Equal(t, resp.Choices[0].Message.FunctionCall.Name, "split_word")

		expectedStruct := counter{
			Count: 7,
			Words: []string{"What", "is", "the", "weather", "like", "in", "Boston?"},
		}
		parser := gollum.NewJSONParserGeneric[counter](false)
		input, err := parser.Parse(ctx, []byte(resp.Choices[0].Message.FunctionCall.Arguments))
		assert.NoError(t, err)
		assert.Equal(t, expectedStruct, input)
	})

	t.Run("callOpenAI", func(t *testing.T) {
		fi := gollum.StructToJsonSchema("ChatCompletion", "Call the OpenAI chat completion API", openai.ChatCompletionRequest{})

		chatRequest := openai.ChatCompletionRequest{
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
			Functions:   []openai.FunctionDefinition{openai.FunctionDefinition(fi)},
		}

		ctx := context.Background()
		resp, err := api.SendRequest(ctx, chatRequest)
		assert.NoError(t, err)
		assert.Equal(t, resp.Model, "gpt-3.5-turbo-0613")
		assert.NotEmpty(t, resp.Choices)
		assert.Empty(t, resp.Choices[0].Message.Content)
		assert.NotNil(t, resp.Choices[0].Message.FunctionCall)
		assert.Equal(t, resp.Choices[0].Message.FunctionCall.Name, "ChatCompletion")

		parser := gollum.NewJSONParserGeneric[openai.ChatCompletionRequest](false)
		input, err := parser.Parse(ctx, []byte(resp.Choices[0].Message.FunctionCall.Arguments))
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

	t.Run("directory", func(t *testing.T) {
		fi := gollum.StructToJsonSchema("directory", "Get the contents of a directory", blobNode{})
		inp := `root
├── dir1
│   ├── file1.txt
│   └── file2.txt
└── dir2
	├── file3.txt
	└── subfolder
		└── file4.txt`

		chatRequest := openai.ChatCompletionRequest{
			Model: "gpt-3.5-turbo-0613",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "user",
					Content: inp,
				},
			},
			MaxTokens:   256,
			Temperature: 0.0,
			Functions:   []openai.FunctionDefinition{openai.FunctionDefinition(fi)},
		}
		ctx := context.Background()
		resp, err := api.SendRequest(ctx, chatRequest)
		assert.NoError(t, err)
		t.Log(resp)
		assert.Equal(t, 0, 1)

		parser := gollum.NewJSONParserGeneric[blobNode](false)
		input, err := parser.Parse(ctx, []byte(resp.Choices[0].Message.FunctionCall.Arguments))
		assert.NoError(t, err)
		assert.NotEmpty(t, input)
		assert.Equal(t, input, blobNode{
			Name: "root",
			Children: []blobNode{
				{
					Name: "dir1",
					Children: []blobNode{
						{
							Name:     "file1.txt",
							NodeType: "file",
						},
						{
							Name:     "file2.txt",
							NodeType: "file",
						},
					},
					NodeType: "folder",
				},
				{
					Name: "dir2",
					Children: []blobNode{
						{
							Name:     "file3.txt",
							NodeType: "file",
						},
						{
							Name: "subfolder",
							Children: []blobNode{
								{
									Name:     "file4.txt",
									NodeType: "file",
								},
							},
							NodeType: "folder",
						},
					},
					NodeType: "folder",
				},
			},
			NodeType: "folder",
		})
	})
	t.Run("planner", func(t *testing.T) {
		fi := gollum.StructToJsonSchema("queryPlanner", "Plan a multi-step query", queryNode{})
		// inp := `Jason is from Canada`

		chatRequest := openai.ChatCompletionRequest{
			Model: "gpt-3.5-turbo-0613",
			Messages: []openai.ChatCompletionMessage{
				{
					Role: "system",
					Content: `When a user asks a question, you must use the 'queryPlanner' function to answer the question. If you are at all unsure, break the question into multiple smaller questions
Example:
Input: What is the population of Jason's home country?

Output:  What is the population of Jason's home country?
│   ├── What is Jason's home country? 
│   ├── What is the population of that country?`,
				},
				{
					Role:    "user",
					Content: "What's on the flag of Jason's home country?",
				},
			},
			MaxTokens:   256,
			Temperature: 0.0,
			Functions:   []openai.FunctionDefinition{openai.FunctionDefinition(fi)},
		}
		ctx := context.Background()
		resp, err := api.SendRequest(ctx, chatRequest)
		assert.NoError(t, err)
		t.Log(resp)
		assert.Equal(t, 0, 1)

	})
}

func BenchmarkStructToJsonSchem(b *testing.B) {
	b.Run("basic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gollum.StructToJsonSchema("queryPlanner", "Plan a multi-step query", queryNode{})
		}
	})
	b.Run("generic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gollum.StructToJsonSchemaGeneric[queryNode]("queryPlanner", "Plan a multi-step query")
		}
	})
}
