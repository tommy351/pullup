package golden

import "sigs.k8s.io/yaml"

type Serializer interface {
	Serialize(interface{}) ([]byte, error)
}

type YAMLSerializer struct{}

func (*YAMLSerializer) Serialize(data interface{}) ([]byte, error) {
	return yaml.Marshal(data)
}
