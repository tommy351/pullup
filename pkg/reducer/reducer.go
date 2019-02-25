package reducer

import "github.com/tommy351/pullup/pkg/model"

type Reducer interface {
	Reduce(resource *model.Resource) error
}
