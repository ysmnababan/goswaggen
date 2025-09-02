package config

var YamlConfigTemplate string = `error_response: "default.Error"
success_response: "default.Success"`
var YamlConfigName string = "goswaggen.yaml"

type Config struct {
	DefaultSuccessResponse string `yaml:"success_response"`
	DefaultFailureResponse string `yaml:"error_response"`
}

var Cfg = &Config{
	DefaultSuccessResponse: "default.Success",
	DefaultFailureResponse: "default.Failure",
}
