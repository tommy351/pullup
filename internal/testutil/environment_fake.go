// +build !integration

package testutil

import (
	"github.com/tommy351/pullup/internal/testenv"
)

func NewEnvironment() testenv.Interface {
	return &testenv.Fake{Scheme: Scheme}
}
