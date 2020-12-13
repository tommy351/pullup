package hookutil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

var ErrInvalidAction = errors.New("invalid action")

type JSONSchemaValidationErrors []gojsonschema.ResultError

func (v JSONSchemaValidationErrors) Error() string {
	details := make([]string, len(v))

	for i, e := range v {
		details[i] = "- " + e.Description()
	}

	return "validation errors:\n" + strings.Join(details, "\n")
}

type ValidationErrors []string

func (v ValidationErrors) Error() string {
	details := make([]string, len(v))

	for i, e := range v {
		details[i] = "- " + e
	}

	return "validation errors:\n" + strings.Join(details, "\n")
}

type JSONSchemaValidateError struct {
	err error
}

func (j JSONSchemaValidateError) Error() string {
	return fmt.Sprintf("json schema validate failed: %s", j.err.Error())
}

func (j JSONSchemaValidateError) Unwrap() error {
	return j.err
}
