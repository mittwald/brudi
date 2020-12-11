package internal

func Unique(stringSlice []string) []string {
	exists := make(map[string]struct{})
	var result []string
	for _, val := range stringSlice {
		if _, ok := exists[val]; !ok {
			exists[val] = struct{}{}
			result = append(result, val)
		}
	}
	return result
}
