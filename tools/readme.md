# tools

This is an implementation of a Tools interface for LLM's to interact with. Tools are entirely arbitrary and can do whatever you want, and just have a very simple API:

```go
type Tool interface {
	Name() string
	Description() string
	Run(ctx context.Context, input interface{}) (interface{}, error)
}
```

A simple implementation would then be something like 

```go
type CalculatorInput struct {
	Expression  string                 `json:"expression" jsonschema:"required" jsonschema_description:"mathematical expression to evaluate"`
	Environment map[string]interface{} `json:"environment,omitempty" jsonschema_description:"optional environment variables to use when evaluating the expression"`
}

type CalculatorTool struct{}

// Run evaluates a mathematical expression and returns it as a string.
func (c *CalculatorTool) Run(ctx context.Context, input interface{}) (interface{}, error) {
    // do stuff
}
```

Note that the input struct has a lot of JSON Schema mappings -- this is so that we can feed the description directly to OpenAI!

## Pipelines

This is very much WIP -- the general idea is that we've defined a simple Reader / Writer interface

```go
type ToolWriter interface {
	Write(ctx context.Context, input interface{}) error
}

type ToolReader interface {
	Read(ctx context.Context) (interface{}, error)
}
```

And then we can chain these together to create a pipeline. For example, we could have a `ToolWriter` that does one LLM lookup and then writes the result to a `ToolReader` that does another LLM lookup. This is useful for chaining together multiple LLMs, or for doing things like "fill in the blank" style completion. This pattern is also useful for concurrent execution, as we can have multiple `ToolWriters` writing to the same `ToolReader` and then have the `ToolReader` do something with the results. An example of this would be performing multiple search requests in parallel and then summarizing the results.