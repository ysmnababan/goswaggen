package testutil

import "path/filepath"

func Ascend(path string, levels int) string {
	for range levels {
		path = filepath.Dir(path)
	}

	return path
}
