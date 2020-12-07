package gitutil

import (
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("ParseRef", func(input string, expected Ref, ok bool) {
	actual, actualOK := ParseRef(input)
	Expect(actual).To(Equal(expected))
	Expect(actualOK).To(Equal(ok))
},
	Entry("branch", "refs/heads/test", Ref{Type: RefTypeBranch, Name: "test"}, true),
	Entry("branch with slash", "refs/heads/feat/something", Ref{Type: RefTypeBranch, Name: "feat/something"}, true),
	Entry("tag", "refs/tags/v0.12", Ref{Type: RefTypeTag, Name: "v0.12"}, true),
	Entry("empty string", "", Ref{}, false),
	Entry("some random thing", "foo", Ref{}, false),
	Entry("not enough slashes", "foo/bar", Ref{}, false),
	Entry("without refs prefix", "a/b/c", Ref{}, false),
)
