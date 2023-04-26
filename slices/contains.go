package slices

func Contains[T comparable](value T, arr []T) bool {
	for _, v := range arr {
		if value == v {
			return true
		}
	}

	return false
}
