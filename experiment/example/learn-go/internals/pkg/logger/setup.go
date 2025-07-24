package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger() {
	// Set time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Environment-aware output
	if os.Getenv("ENV") == "DEVELOPMENT" {
		// Pretty logging for local dev
		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.Kitchen, // e.g. 3:04PM
		}).With().
			Timestamp().
			Caller(). // includes file:line
			Logger()
	} else {
		// JSON structured logs for prod
		log.Logger = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Caller(). // optional
			Logger()
	}
}
