package generator

import "github.com/ysmnababan/goswaggen/experiment/generator/model"

type Parser interface {
	GetFuncName() string
	GetMethod() string
	GetFrameworkName() string
	GetFullPath() string
	GetPayloadInfos() []*model.PayloadInfo
	ReturnResponses() []*model.ReturnResponse
}
