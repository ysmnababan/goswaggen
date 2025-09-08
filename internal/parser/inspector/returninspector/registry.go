package returninspector

import (
	"go/types"

	"github.com/ysmnababan/goswaggen/internal/config"
	"github.com/ysmnababan/goswaggen/internal/parser/inspector/returninspector/echo"
)

func Register(ti *types.Info, cfg *config.Config) []ReturnProcessor {
	echoReturnProcessor := echo.NewReturnProcessor(ti, cfg)
	ret := []ReturnProcessor{echoReturnProcessor}
	return ret
}
