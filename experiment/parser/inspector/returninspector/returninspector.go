package returninspector

import (
	"fmt"
	"go/ast"
	"parser/context"
	"parser/model"
)

type ReturnProcessor interface {
	Name() string
	Match(ast.Node) bool
	Process(ast.Node) *model.ReturnResponse
}

type ReturnInspector struct {
	processors []ReturnProcessor
	Result     []*model.ReturnResponse
}

func NewReturnInspector(hc context.HandlerContext) *ReturnInspector {
	return &ReturnInspector{
		Result:     []*model.ReturnResponse{},
		processors: Register(hc),
	}
}

func (ri *ReturnInspector) Inspect(n ast.Node) {
	for _, p := range ri.processors {
		retResponse := p.Process(n)
		if retResponse != nil {
			ri.Result = append(ri.Result, retResponse)
		}
	}
}

func (ri *ReturnInspector) PrintResult() {
	fmt.Println("Total return statements: ", len(ri.Result))
	for _, val := range ri.Result {
		successTag := "Failure"
		if val.IsSuccess {
			successTag = "Success"
		}
		fmt.Printf(`
// @Accept %s
// @%s %d _%s_ %s`, val.AcceptType, successTag, val.StatusCode, val.SchemaType, val.StructType)
		fmt.Print("\n\n")
	}
}
