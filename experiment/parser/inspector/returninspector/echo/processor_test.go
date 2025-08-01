package echo

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T, input string) (*ast.File, *types.Info) {
	// We need a specific fileset in this test below for positions.
	// Cannot use typecheck helper.
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", input, 0)
	require.NoError(t, err)
	// f := mustParse(fset, input)

	// Type-check the package.
	// We create an empty map for each kind of input
	// we're interested in, and Check populates them.
	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	var conf types.Config
	_, err = conf.Check("fib", fset, []*ast.File{f}, &info)
	require.NoError(t, err)
	return f, &info
}

func TestIsErrorIfStmt(t *testing.T) {
	const input = `
	package fib

	type S string

	var a, b, c = len(b), S(c), "hello"

	func fib(x int) int {
		var err error
		if err != nil {
			return 10
		}
		if x < 2 {
			return x
		}
		if err != nil {
			return 10
		}
		if err == nil {
			return 10
		}
		return fib(x-1) - fib(x-2)
	}`

	f, info := setup(t, input)
	p := EchoReturnProcessor{typesInfo: info}
	ifStmt := []*ast.IfStmt{}
	ast.Inspect(f, func(n ast.Node) bool {
		if t, ok := n.(*ast.IfStmt); ok {
			ifStmt = append(ifStmt, t)
			return false
		}
		return true
	})
	assert.Equal(t, 4, len(ifStmt))
	errorIfStmtCounter := 0
	for _, stmt := range ifStmt {
		if p.isErrorIfStmt(stmt) {
			errorIfStmtCounter++
		}
	}
	assert.Equal(t, 2, errorIfStmtCounter)
}
