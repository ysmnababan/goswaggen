package injector

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
)

// https://eli.thegreenplace.net/2021/rewriting-go-source-code-with-ast-tooling/
// https://www.zakariaamine.com/2022-09-22/ast-package-generate-function-comments/

type injector struct {
	file    *ast.File
	fset    *token.FileSet
	funcAst *ast.FuncDecl
}

func NewInjector(fset *token.FileSet, f *ast.File, fun *ast.FuncDecl) *injector {
	return &injector{
		file:    f,
		fset:    fset,
		funcAst: fun,
	}
}

func (i *injector) insertToCommentGroup(cg *ast.CommentGroup) {
	if len(i.file.Comments) == 0 {
		i.file.Comments = append(i.file.Comments, cg)
		return
	}
	funcLinePos := i.fset.Position(i.funcAst.Pos()).Line

	found := false
	iter := 0
	for iter < len(i.file.Comments) && !found {
		cLinePos := i.fset.Position(i.file.Comments[iter].End()).Line + 1
		fmt.Println(cLinePos)
		if cLinePos > funcLinePos {
			found = true
			break
		}
		iter++
	}
	if !found {
		i.file.Comments = append(i.file.Comments, cg)
		return
	}
	out := append([]*ast.CommentGroup{}, i.file.Comments[:iter]...)
	out = append(out, cg)
	out = append(out, i.file.Comments[iter:]...)
	i.file.Comments = out
}

func (i *injector) InjectComment(comments []string, srcFile io.Writer) error {
	if len(comments) == 0 {
		return errors.New("comments can't be empty")
	}
	astComments := createASTComment(comments, i.funcAst.Pos())
	if i.funcAst.Doc == nil {
		// insert new comment

		blank := &ast.Comment{
			Text:  "//",
			Slash: token.Pos(i.funcAst.Pos() - 1),
		}
		newList := append([]*ast.Comment{blank}, astComments...)
		newCommentGroup := &ast.CommentGroup{
			List: newList,
		}
		// i.file.Comments = append(i.file.Comments, newCommentGroup)
		i.insertToCommentGroup(newCommentGroup)
	} else {
		i.funcAst.Doc.List = astComments
	}

	err := format.Node(srcFile, i.fset, i.file)
	if err != nil {
		return fmt.Errorf("error writing to a file :%w", err)
	}
	return nil
}

func createASTComment(comments []string, pos token.Pos) []*ast.Comment {
	List := []*ast.Comment{}
	for _, c := range comments {
		astComment := &ast.Comment{
			Text: c,

			// make sure the comment pos is positioned before the func `Pos()`.
			// As long as the placement is in between the previous node
			// and the func, it will placed properly
			Slash: token.Pos(pos - 1),
		}
		List = append(List, astComment)
	}
	return List
}
