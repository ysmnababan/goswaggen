package main

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/packages"
)

func main() {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles|
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Dir:  "../example/learn-go/internals", // relative to where you run `go run`
		Fset: token.NewFileSet(),
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		fmt.Println(pkg.PkgPath)
	}
}
