package template

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig"
	"golang.org/x/xerrors"
)

func Render(text string, data interface{}) (string, error) {
	tmpl, err := template.New("").Funcs(sprig.TxtFuncMap()).Parse(text)

	if err != nil {
		return "", xerrors.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", xerrors.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
