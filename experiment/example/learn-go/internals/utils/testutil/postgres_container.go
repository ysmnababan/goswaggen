package testutil

import (
	"context"
	"log"
	"path/filepath"
	"runtime"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const SMTP_SENDER = "user@example.com"

type PostgresTestContainer struct {
	ctr *postgres.PostgresContainer
	dsn string
}

type PostgresContainerConfig struct {
	Database string
	Username string
	Password string
	Image    string
	Scripts  []string // path to the init Scripts
}

type Option func(*PostgresContainerConfig)

func WithScripts(Script ...string) Option {
	return func(c *PostgresContainerConfig) {
		c.Scripts = Script
	}
}

func getMigrationsDir() string {
	_, filename, _, _ := runtime.Caller(0) // this file
	return filepath.Join(filepath.Dir(filename), "../../migrations")
}

func StartPostgresContainer(ctx context.Context, opts ...Option) (*PostgresTestContainer, error) {
	migrationsDir := getMigrationsDir()

	cfg := &PostgresContainerConfig{
		Database: "test-db",
		Username: "postgres",
		Password: "postgres",
		Image:    "postgres:17",
		Scripts: []string{
			filepath.Join(migrationsDir, "00_create_roles.sql"),
			filepath.Join(migrationsDir, "func_trigger.sql"),
			filepath.Join(migrationsDir, "init.sql"),
		},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	ctr, err := postgres.Run(
		ctx,
		cfg.Image,
		postgres.WithInitScripts(cfg.Scripts...),
		postgres.WithDatabase(cfg.Database),
		postgres.WithUsername(cfg.Username),
		postgres.WithPassword(cfg.Password),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return nil, err
	}
	err = ctr.Snapshot(ctx)
	if err != nil {
		log.Printf("failed to create snapshot: %s", err)
		return nil, err
	}

	dbURL, err := ctr.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return nil, err
	}
	return &PostgresTestContainer{
		dsn: dbURL,
		ctr: ctr,
	}, nil
}

func (p *PostgresTestContainer) DSN() string {
	return p.dsn
}

func (p *PostgresTestContainer) Reset(ctx context.Context) error {
	return p.ctr.Restore(ctx)
}

func (p *PostgresTestContainer) Terminate() error {
	return testcontainers.TerminateContainer(p.ctr)
}
