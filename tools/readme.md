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