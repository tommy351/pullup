package jsonschema

import (
	"github.com/xeipuuv/gojsonschema"
	"k8s.io/apimachinery/pkg/api/validation"
)

// nolint: gochecknoinits
func init() {
	gojsonschema.FormatCheckers.Add("kubernetes_name", KubernetesNameFormatChecker{})
	gojsonschema.FormatCheckers.Add("kubernetes_namespace", KubernetesNamespaceFormatChecker{})
}

type KubernetesNamespaceFormatChecker struct{}

func (KubernetesNamespaceFormatChecker) IsFormat(input interface{}) bool {
	s, ok := input.(string)
	if !ok {
		return false
	}

	if len(validation.ValidateNamespaceName(s, false)) > 0 {
		return false
	}

	return true
}

type KubernetesNameFormatChecker struct{}

func (KubernetesNameFormatChecker) IsFormat(input interface{}) bool {
	s, ok := input.(string)
	if !ok {
		return false
	}

	if len(validation.NameIsDNSSubdomain(s, false)) > 0 {
		return false
	}

	return true
}
