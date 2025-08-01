package returninspector

import (
	"parser/context"
	"parser/inspector/returninspector/echo"
)

func Register(hc context.HandlerContext) []ReturnProcessor {
	echoReturnProcessor := echo.NewReturnInspector(hc)
	ret := []ReturnProcessor{echoReturnProcessor}
	return ret
}
