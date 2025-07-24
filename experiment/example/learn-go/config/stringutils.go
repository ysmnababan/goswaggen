package config

func PriorityString(s ...string) string {
	for _, str := range s {
		if str != "" {
			return str
		}
	}
	return ""
}

func PriorityInt(i ...int) int {
	for _, str := range i {
		if str != 0 {
			return str
		}
	}
	return 0
}

func PriorityArrayString(s ...[]string) []string {
	var result []string
	for _, str := range s {
		if len(str) != 0 {
			return str
		}
	}
	return result
}

func UniqueStrings(strings []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, s := range strings {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	return result
}
