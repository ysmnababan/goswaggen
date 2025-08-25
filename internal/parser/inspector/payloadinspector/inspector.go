package payloadinspector

import (
	"fmt"
	"go/ast"
	

	"github.com/ysmnababan/goswaggen/internal/model"
	"github.com/ysmnababan/goswaggen/internal/parser/context"

	
)

type PayloadProcessor interface {
	Match(ast.Node) bool
	Process(ast.Node) *model.PayloadInfo
}

type PayloadInspector struct {
	processors []PayloadProcessor
	Results    []*model.PayloadInfo
}

func NewPayloadInspector(hc context.HandlerContext) *PayloadInspector {
	return &PayloadInspector{
		processors: Register(hc),
		Results:    []*model.PayloadInfo{},
	}
}

func (pi *PayloadInspector) PrintResult() {
	for _, val := range pi.Results {
		if val.BindMethod == "Bind" {
			fmt.Println("Bind:", val.ParamTypes)
			for _, field := range val.FieldLists {
				fmt.Println(*field)
			}
			fmt.Println()
		} else {
			fmt.Printf("%s(%s)\n", val.BindMethod, val.BasicLit)
		}
	}
}

func (pi *PayloadInspector) Inspect(n ast.Node) {
	for _, p := range pi.processors {
		if p.Match(n) {
			ret := p.Process(n)
			if ret != nil {
				pi.Results = append(pi.Results, ret)
			}
		}
	}
}
