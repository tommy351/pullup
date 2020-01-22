package template

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig"
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
