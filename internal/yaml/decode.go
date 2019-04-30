package yaml

import (
	"bytes"
	"io"
	"os"

	"github.com/tommy351/pullup/internal/testutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func DecodeFile(path string) ([]runtime.Object, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	stat, err := file.Stat()

	if err != nil {
		return nil, err
	}

	return Decode(file, int(stat.Size()))
}

func Decode(input io.ReadCloser, size int) ([]runtime.Object, error) {
	var output []runtime.Object
	reader := yaml.NewDocumentDecoder(input)
	defer reader.Close()

	buf := make([]byte, size)

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
		typed, err := testutil.Scheme.New(gvk)

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
