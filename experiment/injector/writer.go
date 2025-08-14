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
	// printSelf()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "target.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	commentMap := ast.NewCommentMap(fset, file, file.Comments)
	ast.Inspect(file, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.FuncDecl:
			cmaps := commentMap[n]
			if len(cmaps) == 0 {
				fmt.Println("no comment found", cmaps == nil)
				newCg := []*ast.CommentGroup{
					{
						List: []*ast.Comment{
							{
								Text:  "// newComment",
								Slash: token.Pos(n.Pos() - 2),
							},
							{
								Text:  "// secondComment",
								Slash: token.Pos(n.Pos() - 1),
							},
						},
					},
				}
				commentMap[n] = newCg
				fmt.Println(newCg[0].List[0].Slash)
				file.Comments = append(file.Comments, newCg[0])
				return true
			}
			for _, cg := range cmaps {
				fmt.Println(cg.Text())
				fmt.Println(len(cg.List))
			}
			cmap := cmaps[0]
			h := cmap.List[1]
			h.Text = "// really?"
			newCmt := []*ast.Comment{cmap.List[1]}
			cmaps[0].List = newCmt
		}
		return true
	})

	srcFile, err := os.Create("target.go")
	if err != nil {
		log.Fatal(err)
	}
	err = format.Node(srcFile, fset, file)
	if err != nil {
		panic(err)
	}
}
