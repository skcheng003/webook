package repository

import (
	"context"
	"github.com/skcheng003/webook/internal/domain"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) (int64, error)
}

type CachedArticleRepository struct {
}
