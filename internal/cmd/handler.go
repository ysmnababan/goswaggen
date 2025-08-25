package cmd

import "github.com/ysmnababan/goswaggen/internal/model"

type IHandler interface {
	GetFuncName() string
	GetMethod() string
	GetFrameworkName() string
	GetFullPath() string
	GetPayloadInfos() []*model.PayloadInfo
	ReturnResponses() []*model.ReturnResponse
}
