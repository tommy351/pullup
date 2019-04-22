package testutil

import (
	"bytes"
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func DecodeYAMLObjects(input io.ReadCloser, bufferSize int) ([]runtime.Object, error) {
	var output []runtime.Object
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

		data := new(unstructured.Unstructured)

		if err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(buf[0:n]), n).Decode(data); err != nil {
			return nil, err
		}

		output = append(output, data)
	}

	return output, nil
}
