package k8s

func StringP(s string) *string {
	return &s
}

func IntP(i int) *int {
	return &i
}

func BoolP(b bool) *bool {
	return &b
}
