package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
)

func RunSpecs(t *testing.T, desc string) {
	var specReporters []ginkgo.Reporter

	if path := os.Getenv("JUNIT_OUTPUT"); path != "" {
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			panic(err)
		}

		specReporters = append(specReporters, reporters.NewJUnitReporter(path))
	}

	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, desc, specReporters)
}
