package util

func ContainsStr(arr []string, target string) bool {
	for _, value := range arr {
		if value == target {
			return true
		}
	}
	return false
}
