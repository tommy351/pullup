package testutil

import (
	"bytes"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func DecodeYAMLFile(path string) ([]runtime.Object, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	stat, err := file.Stat()

	if err != nil {
		return nil, err
	}

	return DecodeYAMLObjects(file, int(stat.Size()))
}

func DecodeYAMLObjects(input io.ReadCloser, bufferSize int) ([]runtime.Object, error) {
	var output []runtime.Object
	scheme := NewScheme()
	reader := yaml.NewDocumentDecoder(input)
	defer reader.Close()

	buf := make([]byte, bufferSize)

	for {
		n, err := reader.Read(buf)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		// Decode data into unstructured
		untyped := new(unstructured.Unstructured)

		if err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(buf[0:n]), n).Decode(untyped); err != nil {
			return nil, err
		}

		// Try to convert the unstructured into typed struct
		gvk := untyped.GroupVersionKind()
		typed, err := scheme.New(gvk)

		if err != nil {
			if runtime.IsNotRegisteredError(err) {
				output = append(output, untyped)
				continue
			}

			return nil, err
		}

		if err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(buf[0:n]), n).Decode(typed); err != nil {
			return nil, err
		}

		output = append(output, typed)
	}

	return output, nil
}
