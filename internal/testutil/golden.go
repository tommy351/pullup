package testutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"sigs.k8s.io/yaml"
)

const golderFileHeader = "# This is a golden file. DO NOT EDIT.\n"

func MatchGolden(path string) types.GomegaMatcher {
	return &goldenMatcher{
		path: path,
	}
}

type goldenMatcher struct {
	path string
}

func (g *goldenMatcher) Match(actual interface{}) (success bool, err error) {
	actualYAML, err := yaml.Marshal(actual)

	if err != nil {
		return false, err
	}

	golden, err := ioutil.ReadFile(g.path)
	shouldUpdate := false

	if err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}

		shouldUpdate = true
	}

	if b, err := strconv.ParseBool(os.Getenv("UPDATE_GOLDEN")); err == nil {
		shouldUpdate = b
	}

	if shouldUpdate {
		actualYAML = append([]byte(golderFileHeader), actualYAML...)
		if err := ioutil.WriteFile(g.path, actualYAML, os.ModePerm); err != nil {
			return false, err
		}

		return true, nil
	}

	return gomega.MatchYAML(golden).Match(actualYAML)
}

func (g *goldenMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto match golden file\n\t%s", actual, g.path)
}

func (g *goldenMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nnot to match golden file\n\t%s", actual, g.path)
}
