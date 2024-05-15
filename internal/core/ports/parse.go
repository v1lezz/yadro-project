package ports

import (
	"context"
	"yadro-project/internal/core/domain"
)

type Parser interface {
	GetCountComicsInServer(ctx context.Context) (int, error)
	PartParse(ctx context.Context, isNotExist []int) ([]domain.Comics, error)
	FullParse(ctx context.Context, cntInServer int) ([]domain.Comics, error)
}
