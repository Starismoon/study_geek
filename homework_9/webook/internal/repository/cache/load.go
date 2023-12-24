package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"time"
)

type LoadCache interface {
	Set(ctx context.Context, val string) error
	Get(ctx context.Context) (int, string, error)
}

type LoadRedisCache struct {
	client     redis.Cmdable
	key        string
	expiration time.Duration
}

func (r *LoadRedisCache) Set(ctx context.Context, val string) error {
	return r.client.Set(ctx, r.key, val, r.expiration).Err()
}

func (r *LoadRedisCache) Get(ctx context.Context) (int, string, error) {
	val, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return 0, "", err
	}
	tmpList := strings.Split(string(val), "_")
	if len(tmpList) != 2 {
		return 0, "", err
	}
	load, err := strconv.Atoi(tmpList[1])
	if err != nil {
		return 0, "", err
	}
	return load, tmpList[0], err
}

func NewLoadRedisCache(client redis.Cmdable) LoadCache {
	return &LoadRedisCache{
		client:     client,
		key:        "load",
		expiration: time.Minute * 3,
	}
}
