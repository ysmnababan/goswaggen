package returninspector

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/ysmnababan/goswaggen/internal/model"
)

type ReturnProcessor interface {
	// Match(ast.Node) bool
	Process(ast.Node) *model.ReturnResponse
}

type ReturnInspector struct {
	processors []ReturnProcessor
	Results    []*model.ReturnResponse
}

func NewReturnInspector(ti *types.Info) *ReturnInspector {
	return &ReturnInspector{
		Results:    []*model.ReturnResponse{},
		processors: Register(ti),
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
// @%s %d _%s_ %s`, val.ProduceType, successTag, val.StatusCode, val.SchemaType, val.ReturnDataType)
		fmt.Print("\n\n")
	}
}
