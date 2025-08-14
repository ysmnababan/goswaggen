package main

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

func (i *injector) InjectComment(comments []string, srcFile io.Writer) error {
	if len(comments) == 0 {
		return errors.New("comments can't be empty")
	}
	cmaps := ast.NewCommentMap(i.fset, i.funcAst, i.file.Comments)
	commentGroups, ok := cmaps[i.funcAst]
	astComments := createASTComment(comments, i.funcAst.Pos())
	if !ok {
		// no comment above function, create new
		newCommentGroup := &ast.CommentGroup{
			List: astComments,
		}
		i.file.Comments = append(i.file.Comments, newCommentGroup)
		// TODO : learn this tomorrow
		i.fset.File(newCommentGroup.End()).AddLine(int(newCommentGroup.End()))
	} else {
		commentGroups[0].List = astComments
	}

	err := format.Node(srcFile, i.fset, i.file)
	if err != nil {
		return fmt.Errorf("error writing to a file :%w", err)
	}
	return nil
}

func createASTComment(comments []string, pos token.Pos) []*ast.Comment {
	List := []*ast.Comment{}
	maxLen := len(comments)
	for i, c := range comments {
		newPos := maxLen - i
		astComment := &ast.Comment{
			Text:  c,
			Slash: token.Pos(pos - token.Pos(newPos)),
		}
		List = append(List, astComment)
	}
	return List
}
