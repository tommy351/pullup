package testutil

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/onsi/ginkgo/reporters"
)

func NewJUnitReporter() *reporters.JUnitReporter {
	dir := filepath.Join(GetProjectRoot(), "reports", "junit")

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err)
	}

	return reporters.NewJUnitReporter(filepath.Join(dir, "results.xml"))
}

func GetProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}
