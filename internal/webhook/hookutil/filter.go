package hookutil

import (
	"regexp"
	"strings"

	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
)

func FilterByConditions(conditions []string, input string) bool {
	for _, c := range conditions {
		if strings.HasPrefix(c, "/") && strings.HasSuffix(c, "/") {
			if matched, _ := regexp.MatchString(c[1:len(c)-1], input); matched {
				return true
			}
		} else if c == input {
			return true
		}
	}

	return false
}

func FilterWebhook(filter *v1beta1.EventSourceFilter, input []string) bool {
	if filter == nil {
		return true
	}

	if len(filter.Include) > 0 {
		included := false

		for _, s := range input {
			if FilterByConditions(filter.Include, s) {
				included = true

				break
			}
		}

		if !included {
			return false
		}
	}

	if len(filter.Exclude) > 0 {
		for _, s := range input {
			if FilterByConditions(filter.Exclude, s) {
				return false
			}
		}
	}

	return true
}
