package ports

import (
	"context"
	"errors"
	"time"
	"yadro-project/internal/core/domain"
)

var ( //errors
	ErrIsExist    = errors.New("is exist")
	ErrIsNotExist = errors.New("is not exist")
)

type ComicsRepository interface {
	GetComics(ctx context.Context) ([]domain.Comics, error)
	GetCountComics(ctx context.Context) (int, error)
	GetIDMissingComics(ctx context.Context, cntInServer int) ([]int, error)
	Add(ctx context.Context, comics domain.Comics, id int) error
	Close(ctx context.Context, updateTime time.Time) error
	GetLastFullCheckTime(ctx context.Context) (time.Time, error)
	UpdateLastFullCheckTime(ctx context.Context, updateTime time.Time) error
	GetLastUpdateTime(ctx context.Context) (time.Time, error)
	GetURLComicsByID(ctx context.Context, ID int) (string, error)
}
