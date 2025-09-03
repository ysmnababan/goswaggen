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

func (i *injector) InsertToCommentGroup(cg *ast.CommentGroup) {
	if len(i.file.Comments) == 0 {
		i.file.Comments = append(i.file.Comments, cg)
		return
	}
	out := make([]*ast.CommentGroup, 0, 1+len(i.file.Comments))
	funcLinePos := i.fset.Position(i.funcAst.Pos()).Line

	inserted := false
	for _, c := range i.file.Comments {
		cLinePos := i.fset.Position(c.End()).Line + 1
		if !inserted && cLinePos >= funcLinePos {
			out = append(out, cg)
			inserted = true
		}
		out = append(out, c)
	}

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
		i.InsertToCommentGroup(newCommentGroup)
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
