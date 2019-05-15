package golden

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/onsi/gomega/types"
)

type Options struct {
	Serializer  Serializer
	Transformer Transformer
}

func Match(path string, options Options) types.GomegaMatcher {
	return &goldenMatcher{
		path:        path,
		serializer:  options.Serializer,
		transformer: options.Transformer,
	}
}

func MatchObject(path string) types.GomegaMatcher {
	return Match(path, Options{
		Serializer:  &YAMLSerializer{},
		Transformer: &ObjectTransformer{},
	})
}

type goldenMatcher struct {
	path        string
	serializer  Serializer
	transformer Transformer
}

func (g *goldenMatcher) Match(actual interface{}) (bool, error) {
	actualContent, err := g.getActualContent(actual)

	if err != nil {
		return false, err
	}

	var expected []byte
	shouldUpdate, _ := strconv.ParseBool(os.Getenv("UPDATE_GOLDEN"))

	if !shouldUpdate {
		if expected, err = g.getExpectedContent(); err != nil {
			if !os.IsNotExist(err) {
				return false, err
			}

			shouldUpdate = true
		}
	}

	if shouldUpdate {
		if err := ioutil.WriteFile(g.path, actualContent, os.ModePerm); err != nil {
			return false, err
		}

		return true, nil
	}

	return bytes.Equal(expected, actualContent), nil
}

func (g *goldenMatcher) getExpectedContent() ([]byte, error) {
	return ioutil.ReadFile(g.path)
}

func (g *goldenMatcher) getActualContent(actual interface{}) ([]byte, error) {
	return g.serializer.Serialize(g.transformer.Transform(actual))
}

func (g *goldenMatcher) getMessage(actual interface{}, message string) string {
	expected, err := g.getExpectedContent()

	if err != nil {
		panic(err)
	}

	actualContent, err := g.getActualContent(actual)

	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("Expected %s match the golden file %q\n%s", message, g.path, Diff(string(expected), string(actualContent)))
}

func (g *goldenMatcher) FailureMessage(actual interface{}) string {
	return g.getMessage(actual, "to")
}

func (g *goldenMatcher) NegatedFailureMessage(actual interface{}) string {
	return g.getMessage(actual, "not to")
}
