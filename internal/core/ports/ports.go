package ports

import (
	"context"
	"time"
	"yadro-project/internal/core/domain"
)

type ComicsRepository interface {
	GetComics(ctx context.Context) ([]domain.Comics, error)
	GetCountComics(ctx context.Context) (int, error)
	GetIDMissingComics(ctx context.Context, cntInServer int) ([]int, error)
	Add(ctx context.Context, comics domain.Comics, id int) error
	Save(ctx context.Context, updateTime time.Time) error
	GetLastFullCheckTime(ctx context.Context) (time.Time, error)
	UpdateLastFullCheckTime(ctx context.Context, updateTime time.Time) error
	GetLastUpdateTime(ctx context.Context) (time.Time, error)
	GetURLComicsByID(ctx context.Context, ID int) (string, error)
}
