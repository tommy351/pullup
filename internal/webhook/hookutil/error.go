package hookutil

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

var ErrInvalidAction = errors.New("invalid action")

type ValidationErrors []string

func (v ValidationErrors) Error() string {
	details := make([]string, len(v))

	for i, e := range v {
		details[i] = "- " + e
	}

	return "validation errors:\n" + strings.Join(details, "\n")
}

type TriggerNotFoundError struct {
	key types.NamespacedName
	err error
}

func (t TriggerNotFoundError) Error() string {
	return fmt.Sprintf("trigger not found: %s/%s", t.key.Namespace, t.key.Name)
}

func (t TriggerNotFoundError) Unwrap() error {
	return t.err
}
