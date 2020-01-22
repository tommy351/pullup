package testutil

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func LoadDocuments(path string) ([]map[string]interface{}, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer file.Close()

	stat, err := file.Stat()

	if err != nil {
		return nil, fmt.Errorf("failed to get stat of file: %w", err)
	}

	var output []map[string]interface{}
	reader := yaml.NewDocumentDecoder(file)
	defer reader.Close()

	buf := make([]byte, stat.Size())

	for {
		n, err := reader.Read(buf)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		data := map[string]interface{}{}

		if err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(buf[0:n]), n).Decode(&data); err != nil {
			return nil, fmt.Errorf("failed to decode data to a map: %w", err)
		}

		output = append(output, data)
	}

	return output, nil
}

func LoadObjects(scheme *runtime.Scheme, path string) (output []runtime.Object, err error) {
	docs, err := LoadDocuments(path)

	if err != nil {
		return nil, fmt.Errorf("failed to load documents: %w", err)
	}

	output = make([]runtime.Object, len(docs))

	for i, doc := range docs {
		if output[i], err = ToObject(scheme, doc); err != nil {
			return nil, fmt.Errorf("failed to decode data to an object: %w", err)
		}
	}

	return output, nil
}
