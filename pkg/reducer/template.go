package reducer

import (
	"bytes"
	"text/template"

	"github.com/ansel1/merry"
	"github.com/tommy351/pullup/pkg/model"
)

type Template struct{}

func (Template) Reduce(resource *model.Resource) error {
	return ReduceBytes(resource, func(data []byte, resource *model.Resource) ([]byte, error) {
		tmpl, err := template.New("").Parse(string(data))

		if err != nil {
			return nil, merry.Wrap(err)
		}

		var buf bytes.Buffer

		if err := tmpl.Execute(&buf, resource); err != nil {
			return nil, merry.Wrap(err)
		}

		return buf.Bytes(), nil
	})
}
