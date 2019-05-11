package testutil

import (
	"bytes"
	"io"
	"os"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func LoadObjects(scheme *runtime.Scheme, path string) ([]runtime.Object, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, xerrors.Errorf("failed to open file: %w", err)
	}

	defer file.Close()

	stat, err := file.Stat()

	if err != nil {
		return nil, xerrors.Errorf("failed to get stat of file: %w", err)
	}

	var output []runtime.Object
	reader := yaml.NewDocumentDecoder(file)
	defer reader.Close()

	buf := make([]byte, stat.Size())

	for {
		n, err := reader.Read(buf)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, xerrors.Errorf("failed to read file: %w", err)
		}

		data := map[string]interface{}{}

		if err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(buf[0:n]), n).Decode(&data); err != nil {
			return nil, xerrors.Errorf("failed to decode data to a map: %w", err)
		}

		obj, err := ToObject(scheme, data)

		if err != nil {
			return nil, xerrors.Errorf("failed to decode data to an object: %w", err)
		}

		output = append(output, obj)
	}

	return output, nil
}
