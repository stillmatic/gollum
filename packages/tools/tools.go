package tools

import "context"

// Tool is an incredibly generic interface for a tool that an LLM can use.
type Tool interface {
	Name() string
	Description() string
	Run(ctx context.Context, input interface{}) (interface{}, error)
}
