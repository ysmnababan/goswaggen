package main

import (
	"fmt"
	"go/token"
	"time"

	"golang.org/x/tools/go/packages"
)

func main() {
	fmt.Println("Benchmarking package load...")

	benchmarkLoad("Lightweight Config", getLightweightConfig())
	benchmarkLoad("Full Config", getFullConfig())
}

func benchmarkLoad(name string, cfg *packages.Config) {
	start := time.Now()

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		fmt.Printf("%s: error loading packages: %v\n", name, err)
		return
	}
	elapsed := time.Since(start)

	// fmt.Printf("Top-level packages loaded: %d\n", len(pkgs))

	// for _, pkg := range pkgs {
	// 	fmt.Printf("Package: %s\n", pkg.ID)
	// 	fmt.Println("Direct imports:")
	// 	for path := range pkg.Imports {
	// 		fmt.Println(" -", path)
	// 	}
	// 	fmt.Println()
	// }

	// Track all visited packages
	// seen := map[string]bool{}
	// var walk func(pkg *packages.Package)

	// walk = func(pkg *packages.Package) {
	// 	if seen[pkg.ID] {
	// 		return
	// 	}
	// 	seen[pkg.ID] = true

	// 	fmt.Printf("Package: %s\n", pkg.ID)
	// 	// for _, file := range pkg.Syntax {
	// 	// fmt.Println("  File:", cfg.Fset.Position(file.Pos()).Filename)
	// 	// }

	// 	for _, imp := range pkg.Imports {
	// 		// fmt.Println("  Import:", name, "->", imp.ID)
	// 		walk(imp)
	// 	}
	// }

	// for _, pkg := range pkgs {
	// 	walk(pkg)
	// }

	// fmt.Printf("Total packages (recursive): %d\n", len(seen))

	fmt.Printf("[%s]\n", name)
	fmt.Printf("  Loaded %d packages in %v\n", len(pkgs), elapsed)

	// Optional: print packages loaded
	for _, pkg := range pkgs {
		fmt.Println("  -", pkg.PkgPath)
	}

	for _, imp := range pkgs[0].Imports {
		if imp.Syntax == nil {
			fmt.Printf("%s has NO Syntax\n", imp.PkgPath)
		} else {
			fmt.Printf("%s has Syntax with %d files\n", imp.PkgPath, len(imp.Syntax))
		}
	}
	fmt.Println()
}

func getLightweightConfig() *packages.Config {
	return &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Fset: token.NewFileSet(),
		Dir:  "../example/learn-go", // <- adjust this
	}
}

func getFullConfig() *packages.Config {
	return &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedSyntax,
		Fset: token.NewFileSet(),
		Dir:  "../example/learn-go", // <- adjust this
	}
}
