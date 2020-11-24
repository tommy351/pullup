package golden

import (
	"fmt"
	"io"

	"sigs.k8s.io/yaml"
)

type YAMLSerializer struct{}

func (YAMLSerializer) Serialize(w io.Writer, input interface{}) error {
	data, err := yaml.Marshal(input)
	if err != nil {
		return fmt.Errorf("yaml marshal error: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	return nil
}
