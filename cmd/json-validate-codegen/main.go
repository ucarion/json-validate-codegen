package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/json-validate/json-validate-codegen"
	"github.com/json-validate/json-validate-codegen/typescript"
	"github.com/json-validate/json-validate-go"

	"github.com/urfave/cli"
)

type outputLang int

const (
	outputLangTypeScript outputLang = iota
)

func main() {
	app := cli.NewApp()
	app.Name = "json-validate-codegen"
	app.Usage = "Generate code from JSON Validate schemas"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "lang",
			Value: "typescript",
			Usage: "language to output",
		},
	}

	app.Action = func(c *cli.Context) error {
		var lang outputLang

		switch c.String("lang") {
		case "typescript":
			lang = outputLangTypeScript
		default:
			return fmt.Errorf("unknown lang: %#v", c.String("lang"))
		}

		return run(c.Args(), lang)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(schemaPaths []string, lang outputLang) error {
	schemas := make([]jsonvalidate.SchemaStruct, len(schemaPaths))
	for i, schemaPath := range schemaPaths {
		reader, err := os.Open(schemaPath)
		if err != nil {
			return err
		}

		decoder := json.NewDecoder(reader)
		err = decoder.Decode(&schemas[i])
		if err != nil {
			return err
		}
	}

	// construct a new validator from the given schemas
	registry, err := jsonvalidate.NewRegistry(schemas)
	if err != nil {
		return err
	}

	encoder := codegen.Encoder{
		Out:      os.Stdout,
		Registry: registry,
		Emitter:  &typescript.Emitter{},
	}

	return encoder.Run()
}
