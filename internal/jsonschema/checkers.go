package jsonschema

import (
	"regexp"

	"github.com/xeipuuv/gojsonschema"
)

// dnsLabelPattern is a pattern for DNS labels.
//
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
var dnsLabelPattern = regexp.MustCompile("^[a-z0-9](?:[a-z0-9-.]*[a-z0-9])?$")

// nolint: gochecknoinits
func init() {
	gojsonschema.FormatCheckers.Add("kubernetes_name", DNSLabelFormatChecker{MaxLength: 253})
	gojsonschema.FormatCheckers.Add("kubernetes_namespace", DNSLabelFormatChecker{MaxLength: 63})
}

type DNSLabelFormatChecker struct {
	MaxLength int
}

func (d DNSLabelFormatChecker) IsFormat(input interface{}) bool {
	s, ok := input.(string)
	if !ok {
		return false
	}

	if d.MaxLength > 0 && len(s) > d.MaxLength {
		return false
	}

	if !dnsLabelPattern.MatchString(s) {
		return false
	}

	return true
}
