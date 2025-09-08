package tagparserutil

import "regexp"

func ParseTag(raw string) map[string]string {
	tagMap := make(map[string]string)
	re := regexp.MustCompile(`(\w+):"((?:\\.|[^"\\])*)"`)

	matches := re.FindAllStringSubmatch(raw, -1)
	for _, m := range matches {
		key := m[1]
		val := m[2]
		tagMap[key] = val
	}
	return tagMap
}
