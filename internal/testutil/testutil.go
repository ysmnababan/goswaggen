package testutil

import (
	"github.com/ysmnababan/goswaggen/internal/fileutil"

	"path/filepath"
	"runtime"
)

func GetVendorTestPath() (string, error) {
	_, testFilePath, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFilePath)
	projectRoot, err := fileutil.FindProjectRoot(testDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, "internal", "testutil", "goenv"), nil
}
