package tools

import "context"

// Tool is an incredibly generic interface for a tool that an LLM can use.
type Tool interface {
	Name() string
	Description() string
	Run(ctx context.Context, input interface{}) (interface{}, error)
}

// ToolWriter is an interface for tools that can write to a tool.
// It's similar to Reader/Writer in the io package.
// It's useful for tools that need to be run in a pipeline.
type ToolWriter interface {
	Write(ctx context.Context, input interface{}) error
}

type ToolReader interface {
	Read(ctx context.Context) (interface{}, error)
}
