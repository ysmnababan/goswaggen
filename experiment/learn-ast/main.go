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

	funPositionMap := make(map[int]*ast.FuncDecl)
	commentPositionMap := make(map[int]*ast.CommentGroup)
	ast.Inspect(f, func(n ast.Node) bool {
		com, ok := n.(*ast.CommentGroup)
		if ok {
			var commentBlockEnd int = fset.Position(com.End()).Line
			commentPositionMap[commentBlockEnd] = com
		}

		fun, ok := n.(*ast.FuncDecl)
		if ok {
			if IsEchoHandler(fun) {
				functionPos := fset.Position(fun.Type.Func).Line
				funPositionMap[functionPos] = fun
			}
		}
		return true
	})
	nodes := AssociateFuncAndComment(funPositionMap, commentPositionMap)
	for _, node := range nodes {
		node.PrintFunctionNode()
	}
}

func IsEchoHandler(n *ast.FuncDecl) bool {
	// param is echo context, len is 1
	param := n.Type.Params.List
	if len(param) != 1 {
		return false
	}
	field := n.Type.Params.List[0]
	switch t := field.Type.(type) {
	case *ast.SelectorExpr:
		fmt.Printf("Parsed: %s.%s\n", t.X, t.Sel.Name)
		if ident, ok := t.X.(*ast.Ident); !ok || ident.Name != "echo" {
			return false
		}
		if t.Sel.Name != "Context" {
			return false
		}
	case *ast.StarExpr:
		if se, ok := t.X.(*ast.SelectorExpr); ok {
			// X = "echo", Sel = "Context"
			ident, _ := se.X.(*ast.Ident)
			fmt.Printf("Parsed: *%s.%s\n", ident.Name, se.Sel.Name)
			if ident.Name != "echo" || se.Sel.Name != "Context" {
				return false
			}
		}
	}

	// result param = 1, type is error
	resParam := n.Type.Results.List
	if len(resParam) != 1 {
		return false
	}

	field = resParam[0]
	switch t := field.Type.(type) {
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); !ok || ident.Name != "error" {
			return false
		}
	}
	return true
}

func AssociateFuncAndComment(fpos map[int]*ast.FuncDecl, cpos map[int]*ast.CommentGroup) []*HandlerNode {
	nodes := make([]*HandlerNode, 0, len(fpos))
	for pos, val := range fpos {
		node := &HandlerNode{
			fn: val,
		}
		if comment, ok := cpos[pos-1]; ok {
			node.comment = comment
		}
		nodes = append(nodes, node)
	}
	return nodes
}
