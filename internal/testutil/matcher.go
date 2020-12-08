package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

func HaveHTTPHeader(key, expected string) types.GomegaMatcher {
	return &HTTPHeaderMatcher{
		Key:      key,
		Expected: expected,
	}
}

type HTTPHeaderMatcher struct {
	Key      string
	Expected string
}

func (h *HTTPHeaderMatcher) Match(actual interface{}) (bool, error) {
	var req *http.Response

	switch v := actual.(type) {
	case *http.Response:
		req = v
	case *httptest.ResponseRecorder:
		// nolint: bodyclose
		req = v.Result()
	default:
		// nolint: goerr113
		return false, fmt.Errorf("expected *http.Response, got %T", v)
	}

	return req.Header.Get(h.Key) == h.Expected, nil
}

func (h *HTTPHeaderMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("to have HTTP header %q", h.Key), h.Expected)
}

func (h *HTTPHeaderMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("not to have HTTP header %q", h.Key), h.Expected)
}
