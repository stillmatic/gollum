package gollum

import (
	"errors"
	"reflect"

	"github.com/invopop/jsonschema"
)

type FunctionInput struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  jsonschema.Schema `json:"parameters"`
}

func StructToJsonSchema(functionName string, functionDescription string, inputStruct interface{}) (FunctionInput, error) {
	var functionInput FunctionInput
	t := reflect.TypeOf(inputStruct)
	schema := jsonschema.ReflectFromType(reflect.Type(t))
	inputStructName := t.Name()
	// only get the
	inputProperties, ok := schema.Definitions[inputStructName]
	if !ok {
		return functionInput, errors.New("could not find input struct in schema")
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
	}, nil
}
