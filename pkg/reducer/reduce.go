package reducer

type Interface interface {
	Reduce(input interface{}) (interface{}, error)
}

func Pipe(input interface{}, reducers ...Interface) (interface{}, error) {
	output := input

	for _, r := range reducers {
		var err error

		if output, err = r.Reduce(output); err != nil {
			return nil, err
		}
	}

	return output, nil
}
