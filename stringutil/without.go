package stringutil

// Without removes the string (str) from the list (values) and returns the result
func Without(values []string, str string) []string {
	ret := []string{}
	for _, v := range values {
		if v == str {
			continue
		}
		ret = append(ret, v)
	}

	return ret
}
