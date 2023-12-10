package repository

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository/cache"
)

type TopNRepository interface {
	GetTopN(ctx context.Context) ([]domain.Article, error)
	SetTopN(ctx context.Context, articles []domain.Article) error
}

type CachedTopNRepository struct {
	cache *cache.TopNRedisCache
}

func (c *CachedTopNRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return c.cache.Get(ctx)
}
func (c *CachedTopNRepository) SetTopN(ctx context.Context, articles []domain.Article) error {
	return c.cache.Set(ctx, articles)
}
