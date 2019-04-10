package testutil

import (
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/types"
)

func NewUID() types.UID {
	return types.UID(uuid.Must(uuid.NewRandom()).String())
}
