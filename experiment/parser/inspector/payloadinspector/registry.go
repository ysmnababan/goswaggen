package payloadinspector

import (
	"parser/context"
	"parser/inspector/payloadinspector/echo"
)

func Register(hc context.HandlerContext) []PayloadProcessor {
	echoPayloadProcessor := echo.NewPayloadProcessor(hc)
	return []PayloadProcessor{
		echoPayloadProcessor,
	}
}
