package kubernetes

import (
	"bytes"
	"text/template"
)

type TemplateReducer struct{}

func (t *TemplateReducer) Reduce(data []byte, resource *Resource) ([]byte, error) {
	tmpl, err := template.New("").Parse(string(data))

	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, resource); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
