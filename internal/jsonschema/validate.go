package jsonschema

import (
	"github.com/xeipuuv/gojsonschema"
)

func Validate(schema, document gojsonschema.JSONLoader) (*gojsonschema.Result, error) {
	return gojsonschema.Validate(schema, document)
}
