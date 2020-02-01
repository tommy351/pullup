package golden

import (
	"io"

	"sigs.k8s.io/yaml"
)

type YAMLSerializer struct{}

func (*YAMLSerializer) Serialize(w io.Writer, input interface{}) error {
	data, err := yaml.Marshal(input)

	if err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}
