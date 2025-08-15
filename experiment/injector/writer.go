package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
)

func main() {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "target.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	var fun *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if f, ok := n.(*ast.FuncDecl); ok {
			if f.Name.Name == "SomeFunc" {
				fun = f
				return false
			}
		}
		return true
	})
	injector := NewInjector(fset, file, fun)
	newCmt := []string{
		"// first",
		"// second",
		"// third",
		"// fourth",
	}
	srcFile, _ := os.Create("target.go")
	err = injector.InjectComment(newCmt, srcFile)
	if err != nil {
		log.Fatal(err)
	}
}
