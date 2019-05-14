package github

import (
	"regexp"
	"strings"

	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
)

func filterWebhook(filter *v1alpha1.WebhookFilter, text string) bool {
	if len(filter.Include) > 0 && !filterByConditions(filter.Include, text) {
		return false
	}

	if len(filter.Exclude) > 0 && filterByConditions(filter.Exclude, text) {
		return false
	}

	return true
}

func filterByConditions(conditions []string, text string) bool {
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
