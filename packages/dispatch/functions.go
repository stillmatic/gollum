package dispatch

import (
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/sashabaranov/go-openai"
)

type FunctionInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters"`
}

type OAITool struct {
	// Type is always "function" for now.
	Type     string        `json:"type"`
	Function FunctionInput `json:"function"`
}

func FunctionInputToTool(fi FunctionInput) openai.Tool {
	f_ := openai.FunctionDefinition(fi)
	return openai.Tool{
		Type:     "function",
		Function: &f_,
	}
}

func StructToJsonSchema(functionName string, functionDescription string, inputStruct interface{}) FunctionInput {
	t := reflect.TypeOf(inputStruct)
	schema := jsonschema.ReflectFromType(reflect.Type(t))
	inputStructName := t.Name()
	// only get the single struct we care about
	inputProperties, ok := schema.Definitions[inputStructName]
	if !ok {
		// this should not happen
		panic("could not find input struct in schema")
	}
	parameters := jsonschema.Schema{
		Type:       "object",
		Properties: inputProperties.Properties,
		Required:   inputProperties.Required,
	}
	return FunctionInput{
		Name:        functionName,
		Description: functionDescription,
		Parameters:  parameters,
	}
}

func StructToJsonSchemaGeneric[T any](functionName string, functionDescription string) FunctionInput {
	var tArr [0]T
	tt := reflect.TypeOf(tArr).Elem()
	schema := jsonschema.ReflectFromType(reflect.Type(tt))
	inputStructName := tt.Name()
	// only get the single struct we care about
	inputProperties, ok := schema.Definitions[inputStructName]
	if !ok {
		// this should not happen
		panic("could not find input struct in schema")
	}
	parameters := jsonschema.Schema{
		Type:       "object",
		Properties: inputProperties.Properties,
		Required:   inputProperties.Required,
	}
	return FunctionInput{
		Name:        functionName,
		Description: functionDescription,
		Parameters:  parameters,
	}
}
