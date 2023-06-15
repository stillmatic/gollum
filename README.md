# GOLLuM

Production-grade LLM tooling. At least, in theory -- stuff changes fast so don't expect stability from this library so much as ideas for your own apps.

## Features

- Parses arbitrary Go structs into JSONSchema for OpenAI - and validates when unmarshaling back to your structs
- Highly performant vector store solution with exact search 
    - SIMD acceleration for 10x better perf than naive approach, constant memory usage
    - Drop-in integration with OpenAI and other embedding providers
    - Carefully mocked, tested, and benchmarked.
- Implementation of HyDE (hypothetical documents embeddings) for enhanced retrieval
- MIT License

# Examples

## Parsing

### Simplest

Imagine you have a function `GetWeather` -- 

```go
type getWeatherInput struct {
	Location string `json:"location" jsonschema_description:"The city and state, e.g. San Francisco, CA" jsonschema:"required"`
	Unit     string `json:"unit,omitempty" jsonschema:"enum=celsius,enum=fahrenheit" jsonschema_description:"The unit of temperature"`
}

type getWeatherOutput struct {
    // ...
}

// GetWeather does something, this dosctring is annoying but theoretically possible to get
func GetWeather(ctx context.Context, inp getWeatherInput) (out getWeatherOutput, err error) {
    return out, err
}
```

This is a common pattern for API design, as it is eay to share the `getWeatherInput` struct (well, imagine if it were public). See, for example, the [GRPC service definitions](https://github.com/grpc/grpc-go/blob/master/examples/helloworld/greeter_server/main.go#L43), or the [Connect RPC implementation](https://github.com/bufbuild/connect-go/blob/main/internal/gen/connect/ping/v1/pingv1connect/ping.connect.go#LL155C6-L155C24). This means we can simplify the logic greatly by assuming a single input struct.

Now, we can construct the responses:

```go
type getWeatherInput struct {
	Location string `json:"location" jsonschema_description:"The city and state, e.g. San Francisco, CA" jsonschema:"required"`
	Unit     string `json:"unit,omitempty" jsonschema:"enum=celsius,enum=fahrenheit" jsonschema_description:"The unit of temperature"`
}

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
        MaxTokens:   512,
        Temperature: 0.0,
    },
    Functions: []gollum.FunctionInput{
        fi,
    },
    FunctionCall: "auto",
}

ctx := context.Background()
resp, err := api.SendRequest(ctx, chatRequest)
parser := gollum.NewJSONParser[getWeatherInput](false)
input, err := parser.Parse(ctx, resp.Choices[0].Message.FunctionCall.Arguments)
```

This example steps through all that, end to end. Some of this is 'sort of' pseudo-code, as the OpenAI clients I use haven't implemented support yet for functions, but it should also hopefully show that minimal modifications are necessary to upstream libraries.

It is also possible to go from just the function definition to a fully formed OpenAI FunctionCall. Reflection gives name of the function for free, godoc parsing can get the function description too. I think in practice though that it's fairly unlikely that you need to change the name/description of the function that often, and in practice the inputs change more often. Using this pattern and compiling once makes the most sense to me. 

We should be able to chain the call for the single input and for the ctx + single input case and return it easily. 

### Recursion on arbitrary structs, no definitions

Say you have a struct that has JSON tags defined. 

```go
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
parser := gollum.NewJSONParser[openai.ChatCompletionRequest](false)
input, err := parser.Parse(ctx, resp.Choices[0].Message.FunctionCall.Arguments)
```

On the first try, this yielded the following result:

```json
 {
  "model": "gpt-3.5-turbo",
  "messages": [
    {"role": "system", "content": "You are Kirby, a friendly virtual assistant."},
    {"role": "user", "content": "What is the definition of recursion?"}
  ]
}
```

That's really sick considering that _no_ effort was put into manually creating a new JSON struct, and the original struct didn't have any JSONSchema tags - just JSON serdes comments.