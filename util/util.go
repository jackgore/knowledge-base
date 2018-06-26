package util

// Contains determines if the given string slice contains
// the given string.
func Contains(strs []string, str string) bool {
	for _, val := range strs {
		if val == str {
			return true
		}
	}

	return false
}
