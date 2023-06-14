package gollum

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"
	val "github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// Parser is an interface for parsing strings into structs
// It is threadsafe and can be used concurrently. The underlying validator is threadsafe as well.
type Parser[T any] interface {
	Parse(ctx context.Context, input string) (T, error)
}

// JSONParser is a parser that parses arbitrary JSON structs
// It is threadsafe and can be used concurrently. The underlying validator is threadsafe as well.
type JSONParser[T any] struct {
	validate bool
	schema   *val.Schema
}

// NewJSONParser returns a new JSONParser
// validation is done via jsonschema
func NewJSONParser[T any](validate bool) *JSONParser[T] {
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

	return &JSONParser[T]{
		validate: validate,
		schema:   sch,
	}
}

func (p *JSONParser[T]) Parse(ctx context.Context, input string) (T, error) {
	var t T
	var v interface{}
	// annoying, must pass an interface to the validate function
	// so we have to unmarshal to interface{}
	err := json.Unmarshal([]byte(input), &v)
	if err != nil {
		return t, errors.Wrap(err, "could not unmarshal input json to interface")
	}
	// but also unmarshal to the struct because it's easy to get a type conversion error
	// e.g. if go struct name doesn't match the json name
	err = json.Unmarshal([]byte(input), &t)
	if err != nil {
		return t, errors.Wrap(err, "could not unmarshal input json to struct")
	}
	if p.validate {
		err := p.schema.Validate(v)
		if err != nil {
			return t, errors.Wrap(err, "error validating input json")
		}
	}

	return t, nil
}

// YAMLParser converts strings into YAML structs
// It is threadsafe and can be used concurrently. The underlying validator is threadsafe as well.
type YAMLParser[T any] struct {
	validate bool
	format   string
}

// NewYamlParser returns a new YAMLParser
// If validate is true, the parser will validate the struct using the struct tags
// The underlying struct must have the `validate` tag if validation is enabled, otherwise it will return `InvalidValidationError`
func NewYamlParser[T any](validate bool) *YAMLParser[T] {
	var t T
	format := ""
	if validate {
		format += "type "
		tt := reflect.TypeOf(t)
		format += tt.Name() + " struct {\n"
		for i := 0; i < tt.NumField(); i++ {
			field := tt.Field(i)
			format += field.Name + "`yaml:" + field.Tag.Get("yaml")
			if field.Tag.Get("validate") != "" {
				format += " validate:" + field.Tag.Get("validate") + "`"
			}
			format += "\n"
		}
		format += "}"
	}

	return &YAMLParser[T]{
		validate: validate,
		format:   format,
	}
}

func (p *YAMLParser[T]) Parse(ctx context.Context, input string) (T, error) {
	var t T
	err := yaml.Unmarshal([]byte(input), &t)
	if err != nil {
		return t, err
	}

	return t, nil
}
