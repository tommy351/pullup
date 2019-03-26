package reducer

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestReducer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "reducer")
}
