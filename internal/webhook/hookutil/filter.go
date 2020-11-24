package hookutil

import (
	"regexp"
	"strings"

	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
)

func FilterByConditions(conditions []string, text string) bool {
	for _, c := range conditions {
		if strings.HasPrefix(c, "/") && strings.HasSuffix(c, "/") {
			if matched, _ := regexp.MatchString(c[1:len(c)-1], text); matched {
				return true
			}
		} else if c == text {
			return true
		}
	}

	return false
}

func FilterWebhook(filter *v1beta1.WebhookFilter, text string) bool {
	if filter == nil {
		return true
	}

	if len(filter.Include) > 0 && !FilterByConditions(filter.Include, text) {
		return false
	}

	if len(filter.Exclude) > 0 && FilterByConditions(filter.Exclude, text) {
		return false
	}

	return true
}
