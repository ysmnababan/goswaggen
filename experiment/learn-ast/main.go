package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	fset := token.NewFileSet()
	src := `package foo

import (
	"fmt"
	"time"
	psr "go/parser"
)

func bar(t string) {
	fmt.Println(time.Now())
}`
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, val := range f.Imports {
		fmt.Println(val.Path.Value)
		fmt.Println(val.Name)
	}

	for _, val := range f.Decls {
		fmt.Println(val.Pos(), val.End())
	}
	// ast.Print(fset, f)

	ast.Inspect(f, func(n ast.Node) bool {
		fun, ok := n.(*ast.FuncDecl)
		if ok {
			// ast.Print(fset, fun)
			fmt.Println("here")
			// fmt.Println(fun.Doc)
			fmt.Println(fun.Name.Name)
			fmt.Println(fun.Type.Params.List[0].Type)
			fmt.Println(fun.Type.Params.List[0].Names)
			// fmt.Println(fun.Type.Results.List[0])
		}
		return true
	})
}
