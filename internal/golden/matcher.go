package golden

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/sergi/go-diff/diffmatchpatch"
	"sigs.k8s.io/yaml"
)

const header = "# This is a golden file. DO NOT EDIT.\n"

func Match(path string) types.GomegaMatcher {
	return &goldenMatcher{
		path: path,
	}
}

type goldenMatcher struct {
	path string
}

func (g *goldenMatcher) Match(actual interface{}) (bool, error) {
	actualYAML, err := g.getActualYAML(actual)

	if err != nil {
		return false, err
	}

	var expected []byte
	shouldUpdate, _ := strconv.ParseBool(os.Getenv("UPDATE_GOLDEN"))

	if !shouldUpdate {
		if expected, err = g.getExpectedYAML(); err != nil {
			if !os.IsNotExist(err) {
				return false, err
			}

			shouldUpdate = true
		}
	}

	if shouldUpdate {
		if err := ioutil.WriteFile(g.path, append([]byte(header), actualYAML...), os.ModePerm); err != nil {
			return false, err
		}

		return true, nil
	}

	return gomega.MatchYAML(expected).Match(actualYAML)
}

func (g *goldenMatcher) getExpectedYAML() ([]byte, error) {
	data, err := ioutil.ReadFile(g.path)

	if err != nil {
		return nil, err
	}

	return bytes.TrimPrefix(data, []byte(header)), nil
}

func (g *goldenMatcher) getActualYAML(actual interface{}) ([]byte, error) {
	return yaml.Marshal(actual)
}

func (g *goldenMatcher) getMessage(actual interface{}, message string) string {
	expected, err := g.getExpectedYAML()

	if err != nil {
		panic(err)
	}

	actualYAML, err := g.getActualYAML(actual)

	if err != nil {
		panic(err)
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(expected), string(actualYAML), false)

	return fmt.Sprintf("Expected %s match golden file\n\x1b[0m%s", message, dmp.DiffPrettyText(diffs))
}

func (g *goldenMatcher) FailureMessage(actual interface{}) string {
	return g.getMessage(actual, "to")
}

func (g *goldenMatcher) NegatedFailureMessage(actual interface{}) string {
	return g.getMessage(actual, "not to")
}
