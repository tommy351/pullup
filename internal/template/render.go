package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func Render(text string, data interface{}) (string, error) {
	tmpl, err := template.New("").Funcs(sprig.TxtFuncMap()).Parse(text)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func RenderFromJSON(text string, jsonData extv1.JSON) (string, error) {
	var data interface{}

	if err := json.Unmarshal(jsonData.Raw, &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return Render(text, data)
}
