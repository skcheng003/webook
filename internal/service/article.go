package service

import (
	"context"
	"github.com/skcheng003/webook/internal/domain"
	"github.com/skcheng003/webook/internal/repository"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, articleId int64) error
	GetById(ctx context.Context, id int64) (art domain.Article, err error)
	GetPubById(ctx context.Context, id int64) (art domain.Article, err error)
}

type articleService struct {
	repo repository.CachedArticleRepository
}

func (svc *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	return repo.
}

func (svc *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	// TODO implement me
	panic("implement me")
}

func (svc *articleService) Withdraw(ctx context.Context, uid int64, articleId int64) error {
	// TODO implement me
	panic("implement me")
}

func (svc *articleService) GetById(ctx context.Context, id int64) (art domain.Article, err error) {
	// TODO implement me
	panic("implement me")
}

func (svc *articleService) GetPubById(ctx context.Context, id int64) (art domain.Article, err error) {
	// TODO implement me
	panic("implement me")
}

