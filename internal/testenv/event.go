package testenv

import (
	"reflect"

	"github.com/onsi/gomega/types"
)

type EventData struct {
	Type    string
	Reason  string
	Message string
}

func matchEvent(actual EventData, expected interface{}) bool {
	switch e := expected.(type) {
	case types.GomegaMatcher:
		if ok, _ := e.Match(actual); ok {
			return true
		}

	default:
		if reflect.DeepEqual(actual, e) {
			return true
		}
	}

	return false
}
