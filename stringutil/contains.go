package stringutil

import "strings"

// Contains returns true if the string v is in the slice arr
func Contains(arr []string, v string) bool {
	for _, str := range arr {
		if str == v {
			return true
		}
	}
	return false
}

// Contains returns true if the string v is in the slice arr
func ContainsPrefix(arr []string, v string) bool {
	for _, str := range arr {
		if strings.HasPrefix(str, v) {
			return true
		}
	}
	return false
}
