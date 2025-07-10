package main

import (
	"fmt"
	"go/ast"
)

type HandlerNode struct {
	fn      *ast.FuncDecl
	comment *ast.CommentGroup
}

func (n *HandlerNode) PrintFunctionNode() {
	if n.comment != nil {
		fmt.Println(n.comment.Text())
	}
	fmt.Printf("func %s(c echo.Context) error>>\n", n.fn.Name.Name)
	fmt.Println("---------------")
	fmt.Println()
	fmt.Println()
}
