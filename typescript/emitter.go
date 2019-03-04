package typescript

import (
	"fmt"
	"html/template"
	"io"

	"github.com/json-validate/json-validate-codegen"
)

type arrayArgs struct {
	Name     string
	Elements string
}

type structArgs struct {
	Name       string
	Properties []structArgsProperties
}

type structArgsProperties struct {
	Name string
	Type string
}

type valueArgs struct {
	Name   string
	Values string
}

type variantArgs struct {
	Name       string
	TagName    string
	TagValue   string
	Properties []structArgsProperties
}

type unionArgs struct {
	Name     string
	Variants []string
}

var (
	arrayFmt = template.Must(template.New("array").Parse(`
{{- /* */ -}}
export type {{ .Name }} = {{ .Elements }}[];
`))

	structFmt = template.Must(template.New("struct").Parse(`
{{- /* */ -}}
export interface {{ .Name }} {
{{- range .Properties }}
  {{ .Name }}: {{ .Type }};
{{- end }}
}
`))

	valuesFmt = template.Must(template.New("values").Parse(`
{{- /* */ -}}
export interface {{ .Name }} {
	[key: string]: {{ .Values }};
}
`))

	variantFmt = template.Must(template.New("variant").Parse(`
{{- /* */ -}}
export interface {{ .Name }} {
	{{ .TagName }}: "{{ .TagValue }}";
{{- range .Properties }}
  {{ .Name }}: {{ .Type }};
{{- end }}
}
`))

	unionFmt = template.Must(template.New("union").Parse(`
{{- /* */ -}}
export type {{ .Name }} =
{{- range $index, $variant := .Variants }}
	{{ if $index }}|{{ end }} {{ $variant }}
{{- end }}
`))
)

// Emitter is an Emitter that outputs TypeScript code.
type Emitter struct{}

func (e *Emitter) PrimitiveEmpty() string {
	return "any"
}

func (e *Emitter) PrimitiveNull() string {
	return "null"
}

func (e *Emitter) PrimitiveBoolean() string {
	return "boolean"
}

func (e *Emitter) PrimitiveNumber() string {
	return "number"
}

func (e *Emitter) PrimitiveString() string {
	return "string"
}

func (e *Emitter) EmitArray(out io.Writer, array codegen.Array) (string, error) {
	name := "Default"
	for _, s := range array.Path.Segments {
		if s.Elements {
			name = name + "Element"
		} else if s.Variants {
			name = name + "Variant"
		} else if s.Values {
			name = name + "Value"
		} else {
			name = name + s.Property
		}
	}

	args := arrayArgs{
		Name:     name,
		Elements: array.Elements,
	}

	err := arrayFmt.Execute(out, args)
	return name, err
}

func (e *Emitter) EmitStruct(out io.Writer, strukt codegen.Struct) (string, error) {
	name := "Default"
	for _, s := range strukt.Path.Segments {
		if s.Elements {
			name = name + "Element"
		} else if s.Variants {
			name = name + "Variant"
		} else if s.Values {
			name = name + "Value"
		} else {
			name = name + s.Property
		}
	}

	args := structArgs{
		Name:       name,
		Properties: []structArgsProperties{},
	}

	for key, value := range strukt.RequiredProperties {
		args.Properties = append(args.Properties, structArgsProperties{
			Name: key,
			Type: value,
		})
	}

	for key, value := range strukt.OptionalProperties {
		args.Properties = append(args.Properties, structArgsProperties{
			Name: fmt.Sprintf("%s?", key),
			Type: value,
		})
	}

	err := structFmt.Execute(out, args)
	return name, err
}

func (e *Emitter) EmitValues(out io.Writer, values codegen.Values) (string, error) {
	name := "Default"
	for _, s := range values.Path.Segments {
		if s.Elements {
			name = name + "Element"
		} else if s.Variants {
			name = name + "Variant"
		} else if s.Values {
			name = name + "Value"
		} else {
			name = name + s.Property
		}
	}

	args := valueArgs{
		Name:   name,
		Values: values.Values,
	}

	err := valuesFmt.Execute(out, args)
	return name, err
}

func (e *Emitter) EmitVariant(out io.Writer, variant codegen.Variant) (string, error) {
	name := "Default"
	for _, s := range variant.Path.Segments {
		if s.Elements {
			name = name + "Element"
		} else if s.Variants {
			name = name + "Variant"
		} else if s.Values {
			name = name + "Value"
		} else {
			name = name + s.Property
		}
	}

	args := variantArgs{
		Name:       name,
		TagName:    variant.TagName,
		TagValue:   variant.TagValue,
		Properties: []structArgsProperties{},
	}

	for key, value := range variant.RequiredProperties {
		args.Properties = append(args.Properties, structArgsProperties{
			Name: key,
			Type: value,
		})
	}

	for key, value := range variant.OptionalProperties {
		args.Properties = append(args.Properties, structArgsProperties{
			Name: fmt.Sprintf("%s?", key),
			Type: value,
		})
	}

	err := structFmt.Execute(out, args)
	return name, err
}

func (e *Emitter) EmitUnion(out io.Writer, union codegen.Union) (string, error) {
	name := "Default"
	for _, s := range union.Path.Segments {
		if s.Elements {
			name = name + "Element"
		} else if s.Variants {
			name = name + "Variant"
		} else if s.Values {
			name = name + "Value"
		} else {
			name = name + s.Property
		}
	}

	args := unionArgs{
		Name:     name,
		Variants: union.Variants,
	}

	err := unionFmt.Execute(out, args)
	return name, err
}
