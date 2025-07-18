package main

import (
	"go/ast"
	"go/token"
	"go/types"
)

type HandlerInfo struct {
	// Basic identification
	Name        string // e.g., "handlerTest"
	FullPath    string // Absolute file path where the handler is defined
	PackagePath string // e.g., "github.com/example/project/handler"

	// AST and Type Info
	FuncDecl *ast.FuncDecl    // Full AST node for the handler
	FuncObj  *types.Func      // Resolved function object
	FuncType *types.Signature // Function signature (inputs/outputs)
	CallExp  *ast.CallExpr    // Where the handler is registered

	// File context
	File *ast.File      // The file the handler was declared in
	Fset *token.FileSet // Needed to compute line/col or position

	// Parameters and Results
	ParamTypes  []types.Type // Types of parameters
	ResultTypes []types.Type // Types of return values

	// Middleware Context
	CalledBy CallSite // Where this handler is registered (e.g., e.GET, e.POST)

	// Import context (for resolving dependencies like echo.Context, etc.)
	Imports []*ast.ImportSpec // Imports available in the file

	// Annotations (optional)
	Comments string // Doc comment associated with the handler (if any)
}

type CallSite struct {
	FilePath string // Where the handler is used (e.g., route registration)
	Selector string // e.g., "e.GET"
	Route    string // e.g., "/next-test"
	Method   string // e.g., "GET"
	Line     int    // Line number in source file
}
