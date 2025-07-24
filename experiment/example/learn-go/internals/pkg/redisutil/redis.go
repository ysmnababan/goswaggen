package redisutil

import (
	"log"
	"time"

	"learn-go/config"

	"github.com/gomodule/redigo/redis"
	zlog "github.com/rs/zerolog/log"
)

type Redis struct {
	client *redis.Pool
}

func NewRedis(cfg config.RedisConfig) *Redis {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			if cfg.Password != "" {
				return redis.Dial("tcp", cfg.Address, redis.DialPassword(cfg.Password), redis.DialDatabase(1))
			}
			return redis.Dial("tcp", cfg.Address, redis.DialDatabase(1))
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			if err != nil {
				return err
			}
			return nil
		},
		MaxIdle:         cfg.MaxIdle,
		MaxActive:       cfg.MaxActive,
		IdleTimeout:     time.Duration(cfg.IdleTimeout) * time.Second,
		Wait:            true,
		MaxConnLifetime: time.Duration(cfg.MaxConnLifeTime) * time.Second,
	}
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("PING")
	if err != nil {
		log.Fatal(err)
	}
	zlog.Info().Msg("successfully ping the redis")
	return &Redis{client: pool}
}

func (r *Redis) Do(command string, args ...interface{}) (interface{}, error) {
	conn := r.client.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)
	return conn.Do(command, args...)
}

func (r *Redis) Get(key string) (string, error) {
	conn := r.client.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)
	return redis.String(conn.Do("GET", key))
}

func (r *Redis) Set(key, value interface{}) (string, error) {
	conn := r.client.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)
	return redis.String(conn.Do("SET", key, value))
}

func (r *Redis) SetEX(key string, value interface{}, expire float64) (string, error) {
	conn := r.client.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)
	return redis.String(conn.Do("SETEX", key, expire, value))
}

func (r *Redis) TTL(key string) (int, error) {
	conn := r.client.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)
	return redis.Int(conn.Do("TTL", key))
}

func (r *Redis) Del(key string) (int64, error) {
	conn := r.client.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)
	return redis.Int64(conn.Do("DEL", key))
}

func (r *Redis) Expire(key string, ttl int) (int64, error) {
	conn := r.client.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)
	return redis.Int64(conn.Do("EXPIRE", key, ttl))
}

func (r *Redis) Keys(pattern string) ([]string, error) {
	conn := r.client.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)
	return redis.Strings(conn.Do("KEYS", pattern))
}

func (r *Redis) FlushAll() error {
	conn := r.client.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)
	_, err := conn.Do("FLUSHALL")
	return err
}
