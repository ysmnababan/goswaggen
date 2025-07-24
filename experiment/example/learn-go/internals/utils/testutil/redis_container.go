package testutil

import (
	"context"
	"fmt"
	"learn-go/config"
	"learn-go/internals/pkg/redisutil"
	"log"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

type RedisTestContainer struct {
	ctr   *tcredis.RedisContainer
	cfg   *config.RedisConfig
	redis *redisutil.Redis
}

func StartRedisContainer() (*RedisTestContainer, error) {
	ctx := context.Background()

	ctr, err := tcredis.Run(ctx,
		"redis:7",
		// tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
		// tcredis.WithConfigFile(filepath.Join("testdata", "redis7.conf")),
	)
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return nil, err
	}
	mappedPort, err := ctr.MappedPort(ctx, "6379/tcp")
	if err != nil {
		log.Printf("failed to get mapped port: %s", err)
		return nil, err
	}
	hostIP, err := ctr.Host(ctx)
	if err != nil {
		log.Printf("failed to get host IP: %s", err)
		return nil, err
	}
	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())
	log.Println("hello", uri)

	cfg := &config.RedisConfig{
		Address:         uri,
		Password:        "",
		MaxIdle:         10,
		MaxActive:       100,
		IdleTimeout:     100,
		MaxConnLifeTime: 600,
	}
	redis := redisutil.NewRedis(*cfg)
	return &RedisTestContainer{
		ctr:   ctr,
		cfg:   cfg,
		redis: redis,
	}, nil
}

func (p *RedisTestContainer) Terminate() error {
	return testcontainers.TerminateContainer(p.ctr)
}

func (p *RedisTestContainer) GetRedisClient() *redisutil.Redis {
	return p.redis
}

func (p *RedisTestContainer) FlushAll() error {
	return p.redis.FlushAll()
}
