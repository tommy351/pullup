package kubernetes

import (
	"bytes"
	"text/template"

	"github.com/ansel1/merry"
)

type TemplateReducer struct{}

func (t TemplateReducer) Reduce(resource *Resource) error {
	return ByteReducerFunc(t.reduceBytes).Reduce(resource)
}

func (t TemplateReducer) reduceBytes(data []byte, resource *Resource) ([]byte, error) {
	tmpl, err := template.New("").Parse(string(data))

	if err != nil {
		return nil, merry.Wrap(err)
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, resource); err != nil {
		return nil, merry.Wrap(err)
	}

	return buf.Bytes(), nil
}
