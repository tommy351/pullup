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
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
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

	log.SetLogger(zap.New(
		zap.UseDevMode(true),
		zap.WriteTo(ginkgo.GinkgoWriter),
	))

	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, desc, specReporters)
}
