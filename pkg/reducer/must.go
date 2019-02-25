package reducer

func Must(reducer Reducer, err error) Reducer {
	if err != nil {
		panic(err)
	}

	return reducer
}
