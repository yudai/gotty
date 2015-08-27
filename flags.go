package main

import (
	"errors"
	"reflect"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/fatih/structs"

	"github.com/yudai/gotty/app"
)

type flag struct {
	name        string
	shortName   string
	description string
}

func generateFlags(flags []flag, hint map[string]string) ([]cli.Flag, error) {
	o := structs.New(app.DefaultOptions)

	results := make([]cli.Flag, len(flags))

	for i, flag := range flags {
		fieldName := fieldName(flag.name, hint)

		field, ok := o.FieldOk(fieldName)
		if !ok {
			return nil, errors.New("No such field: " + fieldName)
		}

		flagName := flag.name
		if flag.shortName != "" {
			flagName += ", " + flag.shortName
		}
		envName := "GOTTY_" + strings.ToUpper(strings.Join(strings.Split(flag.name, "-"), "_"))

		switch field.Kind() {
		case reflect.String:
			results[i] = cli.StringFlag{
				Name:   flagName,
				Value:  field.Value().(string),
				Usage:  flag.description,
				EnvVar: envName,
			}
		case reflect.Bool:
			results[i] = cli.BoolFlag{
				Name:   flagName,
				Usage:  flag.description,
				EnvVar: envName,
			}
		case reflect.Int:
			results[i] = cli.IntFlag{
				Name:   flagName,
				Value:  field.Value().(int),
				Usage:  flag.description,
				EnvVar: envName,
			}
		default:
			return nil, errors.New("Unsupported type: " + fieldName)
		}
	}

	return results, nil
}

func applyFlags(
	options *app.Options,
	flags []flag,
	mappingHint map[string]string,
	c *cli.Context,
) {
	o := structs.New(options)
	for _, flag := range flags {
		if c.IsSet(flag.name) {
			field := o.Field(fieldName(flag.name, mappingHint))
			var val interface{}
			switch field.Kind() {
			case reflect.String:
				val = c.String(flag.name)
			case reflect.Bool:
				val = c.Bool(flag.name)
			case reflect.Int:
				val = c.Int(flag.name)
			}
			field.Set(val)
		}
	}
}

func fieldName(name string, hint map[string]string) string {
	if fieldName, ok := hint[name]; ok {
		return fieldName
	}
	nameParts := strings.Split(name, "-")
	for i, part := range nameParts {
		nameParts[i] = strings.ToUpper(part[0:1]) + part[1:]
	}
	return strings.Join(nameParts, "")
}
