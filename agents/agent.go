package agents

import "context"

type Agent interface {
	Name() string
	Description() string
	Run(context.Context, interface{}) (interface{}, error)
}
