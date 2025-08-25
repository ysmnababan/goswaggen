package testutil

import (
	"fmt"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ysmnababan/goswaggen/internal/fileutil"
	"golang.org/x/tools/go/packages"
)

type temporaryTestFile struct {
	tempFile  string
	fset      *token.FileSet
	fileCount int
}

func NewTemporaryTestFile(tmp string) (*temporaryTestFile, error) {
	src, err := GetVendorTestPath()
	if err != nil {
		return nil, err
	}
	err = fileutil.CopyDir(src, tmp)
	if err != nil {
		return nil, err
	}
	return &temporaryTestFile{
		tempFile: tmp,
		fset:     token.NewFileSet(),
	}, nil
}

func (t *temporaryTestFile) GetTempFile() string {
	return t.tempFile
}

func (t *temporaryTestFile) GetFileSet() *token.FileSet {
	return t.fset
}
func (t *temporaryTestFile) AddNewFile(filename, code string) error {
	return os.WriteFile(filepath.Join(t.tempFile, filename), []byte(code), 0644)
}

func (t *temporaryTestFile) AddNewFileInPackage(packageName, filename, code string) error {
	libDir := filepath.Join(t.tempFile, packageName)
	err := os.Mkdir(libDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create new folder %w", err)
	}
	return os.WriteFile(filepath.Join(libDir, filename), []byte(code), 0644)
}

func (t *temporaryTestFile) BuildPackages() ([]*packages.Package, error) {
	// run `go mod tidy`
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = t.tempFile
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error combining output(%s): %w", string(output), err)
	}
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Dir:  t.tempFile, // relative to where you run `go run
		Fset: t.fset,
		Env:  append(os.Environ(), "GO111MODULE=on", "GOFLAGS=-mod=vendor"),
	}
	pkgs, err := packages.Load(cfg, "./...") // load add the package
	if err != nil {
		return nil, fmt.Errorf("package load error: %w", err)
	}
	fileCount := 0
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			return nil, fmt.Errorf("package error: %w", e)
		}
		fileCount += len(pkg.Syntax)
	}
	t.fileCount = fileCount
	return pkgs, nil
}

func (t *temporaryTestFile) FileCount() int {
	return t.fileCount
}
