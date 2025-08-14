package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"strings"
)

func parseFunc(filename, functionname string) (outfile *ast.File, fun *ast.FuncDecl, fset *token.FileSet) {
	fset = token.NewFileSet()
	if file, err := parser.ParseFile(fset, filename, nil, 0); err == nil {
		fmt.Println(file.Comments)
		for _, d := range file.Decls {
			if f, ok := d.(*ast.FuncDecl); ok && f.Name.Name == functionname {
				f.Name.Name = "printSelves"
				fun = f
				outfile = file
				return
			}
		}
	}
	panic("function not found")
}

// comment before
func printSelf() {
	// Parse source file and extract the AST without comments for
	// this function, with position information referring to the
	// file set fset.
	file, funcAST, fset := parseFunc("writer.go", "printSelf")

	// Print the function body into buffer buf.
	// The file set is provided to the printer so that it knows
	// about the original source formatting and can add additional
	// line breaks where they were present in the source.
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, funcAST.Body)

	// Remove braces {} enclosing the function body, unindent,
	// and trim leading and trailing white space.
	s := buf.String()
	s = s[1 : len(s)-1]
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n\t", "\n"))

	// Print the cleaned-up body text to stdout.
	fmt.Println(s)

	srcFile, err := os.Create("output.go")
	if err != nil {
		log.Fatal(err)
	}
	err = format.Node(srcFile, fset, file)
	if err != nil {
		panic(err)
	}
}

func main() {
	printSelf()
}
