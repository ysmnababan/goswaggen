package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"learn-go/config"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"gorm.io/driver/postgres"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	err           error
	dbConnections *gorm.DB
)

// Database ...
type Database interface {
	Open(logprovider string) (*gorm.DB, error)
	DSN() string
}

// Config ...
type Config struct {
	config.DBConfig
}

// logprovider consist of "zerolog" or "std"
func Init(logprovider string) {
	dbcfg := config.Get().DB

	db := Config{
		DBConfig: dbcfg,
	}
	var sqlDB *sql.DB
	if dbConnections, err = db.Open(logprovider); err != nil {
		panic(fmt.Errorf("connection to db, error: %v", err))
	}

	if sqlDB, err = dbConnections.DB(); err != nil {
		panic(fmt.Errorf("connection to db, error: %v", err))
	}

	sqlDB.SetMaxOpenConns(dbcfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(dbcfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Hour)

	if err = sqlDB.Ping(); err != nil {
		panic(fmt.Errorf("connection to db, error: %v", err))
	}

	zlog.Info().Msg("successfully connected to db")
}

// Connection ...
func Connection() *gorm.DB {
	if dbConnections == nil {
		panic("connection is undefined")
	}
	return dbConnections
}

// DSN ...

func (c Config) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		c.Host, c.Username, c.Password, c.Name, c.Port, c.SSLMode, c.TimeZone)
}

// Open ...
func (c Config) Open(logmode string) (*gorm.DB, error) {

	dialector := postgres.Open(c.DSN())

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                 dblogger[logmode],
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

var dblogger = map[string]logger.Interface{
	"silent": logger.Discard,
	"zerolog": logger.New(
		&zerologAdapter{Logger: zlog.With().Str("component", "gorm").Logger()},
		logger.Config{
			SlowThreshold:             time.Second, // Log queries that take longer than 1 second
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		},
	),
	"std": logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             time.Second, // Log queries that take longer than 1 second
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	}),
}

// zerologAdapter adapts zerolog to GORM's logger interface
type zerologAdapter struct {
	Logger zerolog.Logger
}

func (z *zerologAdapter) Printf(format string, args ...interface{}) {
	z.Logger.Debug().Msgf(format, args...)
}
