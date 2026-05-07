package main

import (
	"os"
	"path/filepath"

	"github.com/stretchr/testify/assert/yaml"
	"github.com/ysmnababan/goswaggen/internal/cmd"
	"github.com/ysmnababan/goswaggen/internal/config"
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	loadConfig(filepath.Join(dir, config.YamlConfigName))
}

func loadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg config.Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return
	}
	if cfg.DefaultSuccessResponse != "" {
		config.Cfg.DefaultSuccessResponse = cfg.DefaultSuccessResponse
	}
	if cfg.DefaultFailureResponse != "" {
		config.Cfg.DefaultFailureResponse = cfg.DefaultFailureResponse
	}
	if cfg.Security != "" {
		config.Cfg.Security = cfg.Security
	}
}

func main() {
	cmd.Execute()
}
