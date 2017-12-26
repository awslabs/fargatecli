package util

func Uniq(elements []string) []string {
	seen := map[string]bool{}
	result := []string{}

	for i := range elements {
		if seen[elements[i]] == true {
		} else {
			seen[elements[i]] = true
			result = append(result, elements[i])
		}
	}

	return result
}
