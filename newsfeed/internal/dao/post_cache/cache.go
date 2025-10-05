package post_cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	PostsKeyFormat    = "grpc:%d:posts"    // grpc:<userid>:posts
	NewsfeedKeyFormat = "grpc:%d:newsfeed" // grpc:<userid>:newsfeed
)

type (
	CacheDao struct {
		cfg CacheConfig

		redisCli *redis.Client
	}
	CacheConfig struct {
		Host string
		Port int
		TTL  time.Duration // not used now
	}
)

func New(cfg CacheConfig) (*CacheDao, error) {
	// TODO: validate config

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	redisCli := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := redisCli.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	dao := &CacheDao{
		cfg:      cfg,
		redisCli: redisCli,
	}
	return dao, nil
}
