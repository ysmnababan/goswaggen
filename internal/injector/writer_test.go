package injector

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInjectComment_NoComment(t *testing.T) {
	fset := token.NewFileSet()
	mainFile := `
package main

import "fmt"

type data struct {
}
func SomeFunc() {
	fmt.Println("this is a function")
}

	`
	file, err := parser.ParseFile(fset, "target.go", mainFile, parser.ParseComments)
	require.NoError(t, err)
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
	srcFile := new(bytes.Buffer)
	err = injector.InjectComment(newCmt, srcFile)
	require.NoError(t, err)
	got := srcFile.String()
	want := `package main

import "fmt"

type data struct {
} //
// first
// second
// third
// fourth
func SomeFunc() {
	fmt.Println("this is a function")
}
`
	assert.Equal(t, want, got)
	fmt.Println(got)
}

func TestInjectComment_FewerComment(t *testing.T) {
	fset := token.NewFileSet()
	mainFile := `
package main

import "fmt"

type data struct {
}
//this is comment
func SomeFunc() {
	fmt.Println("this is a function")
}

	`
	file, err := parser.ParseFile(fset, "target.go", mainFile, parser.ParseComments)
	require.NoError(t, err)
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
		"// this",
		"// is",
		"// very",
		"// very",
		"// very",
		"// very",
		"// very",
		"// very",
		"// very",
		"// very",
		"// very",
		"// long",
		"// comment",
	}
	srcFile := new(bytes.Buffer)
	err = injector.InjectComment(newCmt, srcFile)
	require.NoError(t, err)
	got := srcFile.String()
	want := `package main

import "fmt"

type data struct {
}

// first
// second
// third
// fourth
// this
// is
// very
// very
// very
// very
// very
// very
// very
// very
// very
// long
// comment
func SomeFunc() {
	fmt.Println("this is a function")
}
`
	assert.Equal(t, want, got)
	fmt.Println(got)
}
