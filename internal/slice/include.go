package slice

func IncludeString(arr []string, s string) bool {
	for _, v := range arr {
		if s == v {
			return true
		}
	}

	return false
}
