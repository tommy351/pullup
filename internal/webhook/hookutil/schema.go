package hookutil

import (
	"bytes"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v2"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func ValidateJSONSchema(schemaJSON, docJSON *extv1.JSON) (*extv1.JSON, error) {
	if schemaJSON == nil || schemaJSON.Raw == nil {
		return docJSON, nil
	}

	docRaw := []byte("null")

	if docJSON != nil && docJSON.Raw != nil {
		docRaw = docJSON.Raw
	}

	compiler := jsonschema.NewCompiler()
	url := "schema.json"

	if err := compiler.AddResource(url, bytes.NewReader(schemaJSON.Raw)); err != nil {
		return nil, fmt.Errorf("json schema load error: %w", err)
	}

	schema, err := compiler.Compile(url)
	if err != nil {
		return nil, fmt.Errorf("json schema compile error: %w", err)
	}

	if err := schema.Validate(bytes.NewReader(docRaw)); err != nil {
		return nil, fmt.Errorf("json schema validate error: %w", err)
	}

	return &extv1.JSON{Raw: docRaw}, nil
}
