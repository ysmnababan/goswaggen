package returninspector

import (
	"go/types"

	"github.com/ysmnababan/goswaggen/internal/parser/inspector/returninspector/echo"
)

func Register(ti *types.Info) []ReturnProcessor {
	echoReturnProcessor := echo.NewReturnInspector(ti)
	ret := []ReturnProcessor{echoReturnProcessor}
	return ret
}
