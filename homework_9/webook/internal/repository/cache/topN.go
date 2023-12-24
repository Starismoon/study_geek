package cache

import (
	"context"
	"encoding/json"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type TopNCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

func (t *TopNRedisCache) Set(ctx context.Context, arts []domain.Article) error {
	for i := range arts {
		arts[i].Content = arts[i].Abstract()
	}
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return t.client.Set(ctx, t.key, val, t.expiration).Err()
}

func (t *TopNRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	val, err := t.client.Get(ctx, t.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

type TopNRedisCache struct {
	client     redis.Cmdable
	key        string
	expiration time.Duration
}

func NewTopNRedisCache(client redis.Cmdable) RankingCache {
	return &RankingRedisCache{
		client:     client,
		key:        "articles:top_n",
		expiration: time.Minute,
	}
}
