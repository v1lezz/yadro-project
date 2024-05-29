package ports

import (
	"context"
	"time"
)

type Indexer interface {
	GetNumbersOfNMostRelevantComics(ctx context.Context, n int, keywords []string) ([]int, error)
	UpdateIndex(ctx context.Context, id int, keywords []string) error
	Save(ctx context.Context, updateTime time.Time) error
	GetLastUpdateTime(ctx context.Context) (time.Time, error)
	Clear(ctx context.Context) error
}
