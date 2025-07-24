package config

import (
	"learn-go/internals/utils/env"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	cfg  Config
	once sync.Once
)

func Get() Config {
	once.Do(func() {
		cfg = buildConfig()
	})

	return cfg
}

func buildConfig() Config {
	e := env.NewEnv()

	alternateSchema := e.GetString("SCHEME")
	if alternateSchema != "http" && alternateSchema != "https" {
		alternateSchema = "http"
	}

	alternatePort := e.GetInt("PORT")
	if alternatePort <= 0 {
		alternatePort = 5001
	}

	alternateBaseURL := "localhost:" + strconv.Itoa(alternatePort)

	config := Config{
		App: AppConfig{
			Name:     PriorityString(e.GetString("APP"), "mobile-api"),
			Version:  PriorityString(e.GetString("VERSION"), "0.0.1"),
			Port:     PriorityInt(e.GetInt("PORT"), alternatePort),
			Schema:   PriorityString(e.GetString("SCHEME"), alternateSchema),
			URL:      PriorityString(e.GetString("BASE_URL"), alternateBaseURL),
			LogLevel: PriorityString(e.GetString("LOG_LEVEL")),
		},
		DB: DBConfig{
			Host:         PriorityString(e.GetString("DB_HOST"), "localhost"),
			Username:     PriorityString(e.GetString("DB_USER")),
			Password:     PriorityString(e.GetString("DB_PASS")),
			Port:         PriorityString(e.GetString("DB_PORT"), "5423"),
			Name:         PriorityString(e.GetString("DB_NAME"), "default"),
			MaxIdleConns: PriorityInt(e.GetInt("DB_MAX_IDLE_CONNS"), 2),
			MaxOpenConns: PriorityInt(e.GetInt("DB_MAX_OPEN_CONNS"), 0),
			LogLevel:     PriorityString(e.GetString("DB_LOG_LEVEL"), "info"),

			SSLMode:  PriorityString(e.GetString("DB_SSLMODE")),
			TimeZone: PriorityString(e.GetString("DB_TZ"), "Asia/Jakarta"),
		},

		JWT: JWTConfig{
			Key:              PriorityString(e.GetString("JWT_KEY")),
			ExpiredIn:        PriorityInt(e.GetInt("JWT_EXPIRED_IN")),
			RefreshExpiredIn: PriorityInt(e.GetInt("JWT_REFRESH_EXPIRED_IN")),
		},

		Redis: RedisConfig{
			Address:         PriorityString(e.GetString("REDIS_ADDRESS")),
			Password:        PriorityString(e.GetString("REDIS_PASSWORD")),
			MaxIdle:         PriorityInt(e.GetInt("REDIS_MAX_IDLE")),
			MaxActive:       PriorityInt(e.GetInt("REDIS_MAX_ACTIVE")),
			IdleTimeout:     PriorityInt(e.GetInt("REDIS_IDLE_TIMEOUT")),
			MaxConnLifeTime: PriorityInt(e.GetInt("REDIS_MAX_CONN_LIFE_TIME")),
		},

		EnableSwagger: strings.ToLower(PriorityString(e.GetString("ENABLE_SWAGGER"), "false")) == "true",
	}

	return config
}

func Env() string {
	selectedEnv := strings.ToUpper(strings.TrimSpace(os.Getenv("ENV")))
	if selectedEnv != "DEVELOPMENT" && selectedEnv != "STAGING" && selectedEnv != "PRODUCTION" {
		selectedEnv = "LOCAL"
	}
	return selectedEnv
}
