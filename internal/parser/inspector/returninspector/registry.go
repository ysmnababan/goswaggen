package returninspector

import (
	"go/types"

	"github.com/ysmnababan/goswaggen/internal/model"
	"github.com/ysmnababan/goswaggen/internal/parser/inspector/returninspector/echo"
)

func Register(ti *types.Info, cfg *model.Config) []ReturnProcessor {
	echoReturnProcessor := echo.NewReturnProcessor(ti, cfg)
	ret := []ReturnProcessor{echoReturnProcessor}
	return ret
}
