package model

type Config struct {
	DefaultSuccessResponse string
	DefaultFailureResponse string
}

var Cfg = Config{
	DefaultSuccessResponse: "default.Success",
	DefaultFailureResponse: "default.Failure",
}
