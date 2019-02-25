package reducer

import "github.com/tommy351/pullup/pkg/model"

type Func func(resource *model.Resource) error

func (r Func) Reduce(resource *model.Resource) error {
	return r(resource)
}
