package reducer

type DebugLogger func(input, output interface{})

func Debug(reducer Interface, logger DebugLogger) Interface {
	return Func(func(input interface{}) (interface{}, error) {
		output, err := reducer.Reduce(input)

		if err != nil {
			return nil, err
		}

		logger(input, output)
		return output, nil
	})
}

func DebugReducers(reducers Reducers, logger DebugLogger) Interface {
	output := make(Reducers, len(reducers))

	for i, r := range reducers {
		output[i] = Debug(r, logger)
	}

	return output
}
