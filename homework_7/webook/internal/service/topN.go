package service

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"sync"
)

type TopNService interface {
	GetTopN(ctx context.Context) ([]domain.Article, error)
}
type topNService struct {
	repo        repository.TopNRepository
	interRepo   repository.InteractiveRepository
	articleRepo repository.ArticleRepository
	Size        int
	lock        sync.Locker
}

func NewtopNService(repo repository.TopNRepository,
	interRepo repository.InteractiveRepository,
	articleRepo repository.ArticleRepository) TopNService {
	return &topNService{
		repo:        repo,
		interRepo:   interRepo,
		articleRepo: articleRepo,
		Size:        100,
	}

}
func (t *topNService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	articles, err := t.repo.GetTopN(ctx)
	if err == nil {
		return articles, nil
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	// 检测其他线程是否已查询过存入缓存
	articles, err = t.repo.GetTopN(ctx)
	if err == nil {
		return articles, nil
	}

	// 从数据库查找
	interActiveList, err := t.interRepo.GetTopN(ctx, "article", t.Size)
	if err != nil {
		return nil, err
	}
	articles = make([]domain.Article, 0, t.Size)
	for _, v := range interActiveList {
		article, er := t.articleRepo.GetById(ctx, v.BizId)
		if er != nil {
			articles = append(articles, article)
		}
	}
	go func() {
		_ = t.repo.SetTopN(ctx, articles)
	}()
	return articles, nil
}
