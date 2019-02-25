package reducer

import (
	"github.com/ansel1/merry"
	"github.com/tommy351/pullup/pkg/model"
)

type List []Reducer

func (list List) Reduce(resource *model.Resource) error {
	for _, reducer := range list {
		if err := reducer.Reduce(resource); err != nil {
			return merry.Wrap(err)
		}
	}

	return nil
}
