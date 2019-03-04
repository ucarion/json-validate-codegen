package codegen

import (
	"errors"
	"io"
	"net/url"

	"github.com/json-validate/json-validate-go"
)

// Encoder drives code generation.
//
// Encoder works by walking across the schemas in a Registry, and calling out to
// an underlying Emitter to produce code for the particular language at hand.
type Encoder struct {
	// Where Encoder should produce its output.
	Out io.Writer

	// The registry to walk over.
	Registry jsonvalidate.Registry

	// The emitter handling the specifics of the target language.
	Emitter Emitter
}

type NamePath struct {
	SchemaID *url.URL
	Segments []NamePathSegment
}

func (np *NamePath) Push(segment NamePathSegment) {
	np.Segments = append(np.Segments, segment)
}

func (np *NamePath) Pop() {
	np.Segments = np.Segments[:len(np.Segments)-1]
}

type NamePathSegment struct {
	Elements bool
	Values   bool
	Variants bool
	Property string
}

type Struct struct {
	Path               *NamePath
	RequiredProperties map[string]string
	OptionalProperties map[string]string
}

type Array struct {
	Path     *NamePath
	Elements string
}

type Values struct {
	Path   *NamePath
	Values string
}

type Variant struct {
	Path               *NamePath
	TagName            string
	TagValue           string
	RequiredProperties map[string]string
	OptionalProperties map[string]string
}

type Union struct {
	Path     *NamePath
	Variants []string
}

// Emitter handles producing code for a particular target language.
type Emitter interface {
	// PrimitiveEmpty returns the name of the "empty" or "top" type.
	PrimitiveEmpty() string

	// PrimitiveNull returns the name of the "null" type.
	PrimitiveNull() string

	// PrimitiveBoolean returns the name of the "boolean" type.
	PrimitiveBoolean() string

	// PrimitiveNumber returns the name of the "number" or "float64" type.
	PrimitiveNumber() string

	// PrimitiveString returns the name of the "string" type.
	PrimitiveString() string

	// EmitStruct outputs a representation of a struct, returning the name of the
	// emitted struct type.
	EmitStruct(io.Writer, Struct) (string, error)

	// EmitArray outputs a representation of an array, returning the name of the
	// emitted array type.
	EmitArray(io.Writer, Array) (string, error)

	// EmitValues outputs a representation of a dictionary, returning the name of
	// the emitted dictionary type.
	EmitValues(io.Writer, Values) (string, error)

	// EmitVariant outputs a representation of a struct that is a variant of a
	// discriminated union, returning the name of the emitted type.
	EmitVariant(io.Writer, Variant) (string, error)

	// EmitUnion outputs a representation of a discriminated union, returning the
	// name of the emitted type.
	EmitUnion(io.Writer, Union) (string, error)
}

// Run triggers the code generation process.
func (e *Encoder) Run() error {
	for _, schema := range e.Registry.Schemas {
		path := NamePath{SchemaID: schema.ID, Segments: []NamePathSegment{}}
		if _, err := e.walk(&path, schema); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) walk(path *NamePath, schema *jsonvalidate.Schema) (string, error) {
	switch schema.Kind {
	case jsonvalidate.SchemaKindEmpty:
		return e.Emitter.PrimitiveEmpty(), nil
	case jsonvalidate.SchemaKindType:
		switch schema.Type {
		case jsonvalidate.SchemaTypeNull:
			return e.Emitter.PrimitiveNull(), nil
		case jsonvalidate.SchemaTypeBoolean:
			return e.Emitter.PrimitiveBoolean(), nil
		case jsonvalidate.SchemaTypeNumber:
			return e.Emitter.PrimitiveNumber(), nil
		case jsonvalidate.SchemaTypeString:
			return e.Emitter.PrimitiveString(), nil
		}
	case jsonvalidate.SchemaKindElements:
		path.Push(NamePathSegment{Elements: true})
		name, err := e.walk(path, schema.Elements)
		if err != nil {
			return "", err
		}

		path.Pop()

		return e.Emitter.EmitArray(e.Out, Array{
			Path:     path,
			Elements: name,
		})
	case jsonvalidate.SchemaKindProperties:
		required := map[string]string{}
		for key, value := range schema.Properties {
			path.Push(NamePathSegment{Property: key})
			name, err := e.walk(path, value)
			if err != nil {
				return "", err
			}

			required[key] = name
			path.Pop()
		}

		optional := map[string]string{}
		for key, value := range schema.OptionalProperties {
			path.Push(NamePathSegment{Property: key})
			name, err := e.walk(path, value)
			if err != nil {
				return "", err
			}

			optional[key] = name
			path.Pop()
		}

		return e.Emitter.EmitStruct(e.Out, Struct{
			Path:               path,
			RequiredProperties: required,
			OptionalProperties: optional,
		})
	case jsonvalidate.SchemaKindValues:
		path.Push(NamePathSegment{Values: true})
		name, err := e.walk(path, schema.Values)
		if err != nil {
			return "", err
		}

		path.Pop()

		return e.Emitter.EmitValues(e.Out, Values{
			Path:   path,
			Values: name,
		})
	case jsonvalidate.SchemaKindDiscriminator:
		path.Push(NamePathSegment{Variants: true})
		variants := []string{}
		for key, variant := range schema.DiscriminatorMapping {
			// Check that the variant value is of kind "properties", as this is the
			// only format supported (at least for now).
			if variant.Kind != jsonvalidate.SchemaKindProperties {
				return "", errors.New("schemas within `mapping` must use only properties and optionalProperties")
			}

			path.Push(NamePathSegment{Property: key})

			required := map[string]string{}
			for key, value := range schema.Properties {
				path.Push(NamePathSegment{Property: key})
				name, err := e.walk(path, value)
				if err != nil {
					return "", err
				}

				required[key] = name
				path.Pop()
			}

			optional := map[string]string{}
			for key, value := range schema.OptionalProperties {
				path.Push(NamePathSegment{Property: key})
				name, err := e.walk(path, value)
				if err != nil {
					return "", err
				}

				optional[key] = name
				path.Pop()
			}

			name, err := e.Emitter.EmitVariant(e.Out, Variant{
				Path:               path,
				TagName:            schema.DiscriminatorPropertyName,
				TagValue:           key,
				RequiredProperties: required,
				OptionalProperties: optional,
			})

			if err != nil {
				return "", err
			}

			path.Pop()

			variants = append(variants, name)
		}

		path.Pop()

		return e.Emitter.EmitUnion(e.Out, Union{
			Path:     path,
			Variants: variants,
		})
	}

	return "", nil
}
