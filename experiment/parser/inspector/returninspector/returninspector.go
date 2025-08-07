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
	Results    []*model.ReturnResponse
}

func NewReturnInspector(hc context.HandlerContext) *ReturnInspector {
	return &ReturnInspector{
		Results:    []*model.ReturnResponse{},
		processors: Register(hc),
	}
}

func (ri *ReturnInspector) Inspect(n ast.Node) {
	for _, p := range ri.processors {
		retResponse := p.Process(n)
		if retResponse != nil {
			ri.Results = append(ri.Results, retResponse)
		}
	}
}

func (ri *ReturnInspector) PrintResult() {
	fmt.Println("Total return statements: ", len(ri.Results))
	for _, val := range ri.Results {
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
