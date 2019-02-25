package reducer

import (
	"github.com/tommy351/pullup/pkg/model"
)

type If struct {
	Condition func(resource *model.Resource) bool
	True      Reducer
	False     Reducer
}

func (i *If) Reduce(resource *model.Resource) error {
	if i.Condition(resource) {
		if i.True != nil {
			return i.True.Reduce(resource)
		}
	} else {
		if i.False != nil {
			return i.False.Reduce(resource)
		}
	}

	return nil
}
