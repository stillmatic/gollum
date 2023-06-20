package gollum

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"
	val "github.com/santhosh-tekuri/jsonschema/v5"
)

// Parser is an interface for parsing strings into structs
// It is threadsafe and can be used concurrently. The underlying validator is threadsafe as well.
type Parser[T any] interface {
	Parse(ctx context.Context, input []byte) (T, error)
}

// JSONParser is a parser that parses arbitrary JSON structs
// It is threadsafe and can be used concurrently. The underlying validator is threadsafe as well.
type JSONParserGeneric[T any] struct {
	validate bool
	schema   *val.Schema
}

// NewJSONParser returns a new JSONParser
// validation is done via jsonschema
func NewJSONParserGeneric[T any](validate bool) *JSONParserGeneric[T] {
	var sch *val.Schema
	var t_ T
	if validate {
		// reflect T into schema
		t := reflect.TypeOf(t_)
		schema := jsonschema.ReflectFromType(t)
		// compile schema
		b, err := schema.MarshalJSON()
		if err != nil {
			panic(errors.Wrap(err, "could not marshal schema"))
		}
		schemaStr := string(b)
		sch, err = val.CompileString("schema.json", schemaStr)
		if err != nil {
			panic(errors.Wrap(err, "could not compile schema"))
		}
	}

	return &JSONParserGeneric[T]{
		validate: validate,
		schema:   sch,
	}
}

func (p *JSONParserGeneric[T]) Parse(ctx context.Context, input []byte) (T, error) {
	var t T
	// but also unmarshal to the struct because it's easy to get a type conversion error
	// e.g. if go struct name doesn't match the json name
	err := json.Unmarshal(input, &t)
	if err != nil {
		return t, errors.Wrap(err, "could not unmarshal input json to struct")
	}

	// annoying, must pass an interface to the validate function
	// so we have to unmarshal to interface{} twice
	if p.validate {
		var v interface{}
		err := json.Unmarshal(input, &v)
		if err != nil {
			return t, errors.Wrap(err, "could not unmarshal input json to interface")
		}
		err = p.schema.Validate(v)
		if err != nil {
			return t, errors.Wrap(err, "error validating input json")
		}
	}

	return t, nil
}
