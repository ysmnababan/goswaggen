package returninspector

import (
	"github.com/ysmnababan/goswaggen/internal/parser/context"
	"github.com/ysmnababan/goswaggen/internal/parser/inspector/returninspector/echo"

)

func Register(hc context.HandlerContext) []ReturnProcessor {
	echoReturnProcessor := echo.NewReturnInspector(hc)
	ret := []ReturnProcessor{echoReturnProcessor}
	return ret
}
