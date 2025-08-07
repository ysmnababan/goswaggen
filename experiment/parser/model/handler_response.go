package model

import "go/ast"

type ReturnResponse struct {
	ReturnStmt *ast.ReturnStmt
	StructType string
	SchemaType string // {object}, string, int, etc
	StatusCode int
	IsSuccess  bool
	AcceptType string // json, xml, string
	FrameWork  string // echo, gin,
}
