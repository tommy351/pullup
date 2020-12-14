package hookutil

import (
	"github.com/xeipuuv/gojsonschema"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func ValidateJSONSchema(schema, doc *extv1.JSON) (*extv1.JSON, error) {
	if schema == nil || schema.Raw == nil {
		return doc, nil
	}

	docRaw := []byte("null")

	if doc != nil && doc.Raw != nil {
		docRaw = doc.Raw
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema.Raw),
		gojsonschema.NewBytesLoader(docRaw),
	)
	if err != nil {
		return nil, JSONSchemaValidateError{err: err}
	}

	if !result.Valid() {
		return nil, JSONSchemaValidationErrors(result.Errors())
	}

	return &extv1.JSON{Raw: docRaw}, nil
}
