package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var rootDir string

func SetRootDir(r string) {
	rootDir = r
}

func RootDir() string {
	return rootDir
}

func FindProjectRoot(startDir string) (string, error) {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// copyDir recursively copies a directory from src to dst.
func CopyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create the destination path by replacing src prefix with dst
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create the directory if it doesn't exist
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file
		return copyFile(path, destPath, info.Mode())
	})
}

func copyFile(srcFile string, dstFile string, mode os.FileMode) error {
	from, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.OpenFile(dstFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	return err
}
