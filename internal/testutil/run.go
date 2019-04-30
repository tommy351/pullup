package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
)

func RunSpecs(t *testing.T, desc string) {
	var specReporters []ginkgo.Reporter

	if dir := os.Getenv("JUNIT_OUTPUT"); dir != "" {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			panic(err)
		}

		path := filepath.Join(dir, fmt.Sprintf("junit-%d.xml", time.Now().UnixNano()))
		specReporters = append(specReporters, reporters.NewJUnitReporter(path))
	}

	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, desc, specReporters)
}
