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

// this is comment
func bar(t string) {
	fmt.Println(time.Now())
}`
	_ = src

	f, err := parser.ParseFile(
		fset,
		// "",
		"./../example/learn-go/internals/app/example_feat/controller.go",
		nil,
		parser.ParseComments)

	if err != nil {
		fmt.Println(err)
		return
	}
	for _, val := range f.Imports {
		fmt.Println(val.Path.Value)
		name := val.Name
		if name != nil {
			fmt.Println(val.Name)
		}
	}

	// for _, val := range f.Decls {
	// fmt.Println(val)
	// fmt.Println(val.Pos(), val.End())
	// }
	// ast.Print(fset, f)

	var functionPos int
	ast.Inspect(f, func(n ast.Node) bool {
		fun, ok := n.(*ast.FuncDecl)

		if ok && fun.Name.Name == "Login" {
			fmt.Println(fun.Name.Name)
			fmt.Println(fun.Type.Params.List[0].Names[0].Name)
			fmt.Println(fun.Type.Params.List[0].Type)
			fmt.Println(fun.Type.Results.List[0].Names)
			fmt.Println(fun.Type.Results.List[0].Type)
			functionPos = fset.Position(fun.Type.Func).Line
			fmt.Println("function pos:", functionPos)
		}
		com, ok := n.(*ast.CommentGroup)
		if ok {
			var commentBlockEnd int = fset.Position(com.End()).Line
			fmt.Println("comment block end:", commentBlockEnd)
			if functionPos == commentBlockEnd+1 {
				fmt.Println(com.Text())
			}
			// fmt.Println(fset.Position(com.End()).Line)
			// ast.Print(fset, com.List)
		}
		return true
	})
}
