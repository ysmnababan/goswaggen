package env

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Env interface {
	GetString(name string) string
	GetBool(name string) bool
	GetInt(name string) int
	GetFloat(name string) float64
}

type env struct{}

func NewEnv() *env {
	return &env{}
}

func (e *env) Load(env string) {
	cwd, _ := os.Getwd()
	err := godotenv.Load(filepath.Join(cwd, env))
	if err != nil {
		log.Error().Err(err).Str("cwd", cwd).Msg("Load .env file error")
	}
}

func (e *env) GetString(name string) string {
	return os.Getenv(name)
}

func (e *env) GetBool(name string) bool {
	s := e.GetString(name)
	i, err := strconv.ParseBool(s)
	if nil != err {
		return false
	}
	return i
}

func (e *env) GetInt(name string) int {
	s := e.GetString(name)
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func (e *env) GetFloat(name string) float64 {
	s := e.GetString(name)
	i, err := strconv.ParseFloat(s, 64)
	if nil != err {
		return 0
	}
	return i
}
