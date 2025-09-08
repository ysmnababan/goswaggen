package payloadinspector

import (
	"github.com/ysmnababan/goswaggen/internal/parser/context"
	"github.com/ysmnababan/goswaggen/internal/parser/inspector/payloadinspector/echo"
)

func Register(hc context.HandlerContext) []PayloadProcessor {
	echoPayloadProcessor := echo.NewPayloadProcessor(hc)
	return []PayloadProcessor{
		echoPayloadProcessor,
	}
}
