package hookutil

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
)

var _ = DescribeTable("FilterByConditions", func(conditions []string, input string, expected bool) {
	Expect(FilterByConditions(conditions, input)).To(Equal(expected))
},
	Entry("Exact match - true", []string{"abc"}, "abc", true),
	Entry("Exact match - false", []string{"abc"}, "cba", false),
	Entry("Pattern - true", []string{`/^a\db$/`}, "a1b", true),
	Entry("Pattern - false", []string{`/^a\db$/`}, "axb", false),
	Entry("Multiple - true 1", []string{"abc", "xyz"}, "abc", true),
	Entry("Multiple - true 2", []string{"abc", "xyz"}, "xyz", true),
	Entry("Multiple - false", []string{"abc", "xyz"}, "def", false),
)

var _ = Describe("FilterWebhook", func() {
	DescribeTable("include only", func(input []string, expected bool) {
		filter := &v1beta1.WebhookFilter{
			Include: []string{"abc", `/^a\db$/`},
		}
		Expect(FilterWebhook(filter, input)).To(Equal(expected))
	},
		Entry("Exact match - true", []string{"abc"}, true),
		Entry("Exact match - false", []string{"xyz"}, false),
		Entry("Pattern - true", []string{"a1b"}, true),
		Entry("Pattern - false", []string{"axb"}, false),
		Entry("Multi input - true", []string{"a", "ab", "abc"}, true),
		Entry("Multi input - false", []string{"x", "xy", "xyz"}, false),
	)

	DescribeTable("exclude only", func(input []string, expected bool) {
		filter := &v1beta1.WebhookFilter{
			Exclude: []string{"abc", `/^a\db$/`},
		}
		Expect(FilterWebhook(filter, input)).To(Equal(expected))
	},
		Entry("Exact match - true", []string{"xyz"}, true),
		Entry("Exact match - false", []string{"abc"}, false),
		Entry("Pattern - true", []string{"axb"}, true),
		Entry("Pattern - false", []string{"a1b"}, false),
		Entry("Multi input - true", []string{"x", "xy", "xyz"}, true),
		Entry("Multi input - false", []string{"a", "ab", "abc"}, false),
	)

	DescribeTable("include & exclude", func(input []string, expected bool) {
		filter := &v1beta1.WebhookFilter{
			Include: []string{"abc", `/a[bc]/`},
			Exclude: []string{"ac"},
		}
		Expect(FilterWebhook(filter, input)).To(Equal(expected))
	},
		Entry("Exact match include", []string{"abc"}, true),
		Entry("Exact match exclude", []string{"ac"}, false),
		Entry("Pattern match include", []string{"ab"}, true),
		Entry("Multi input include", []string{"a", "ab"}, true),
		Entry("Multi input exclude", []string{"a", "ac"}, false),
		Entry("Multi input include & exclude", []string{"a", "ab", "ac", "abc"}, false),
	)
})
