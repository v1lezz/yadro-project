package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"yadro-project/internal/core/domain"
	"yadro-project/internal/core/ports"
	"yadro-project/pkg/pair"
)

type ComicsService struct {
	repo    ports.ComicsRepository
	parser  ports.Parser
	indexer ports.Indexer
	stemmer ports.Stemmer
}

func NewComicsService(repo ports.ComicsRepository, parser ports.Parser, indexer ports.Indexer, stemmer ports.Stemmer) *ComicsService {
	return &ComicsService{
		repo:    repo,
		parser:  parser,
		indexer: indexer,
		stemmer: stemmer,
	}
}

var (
	ErrContextDone = errors.New("server is not accepting new requests")
)

func (srv *ComicsService) GetComics(ctx context.Context, search string) ([]domain.Comics, error) {
	select {
	case <-ctx.Done():
		return nil, ErrContextDone
	default:
	}
	stemmed, err := srv.stemmer.Stem(strings.FieldsFunc(search, func(r rune) bool {
		return !unicode.IsLetter(r) && r != '\''
	}))
	if err != nil {
		return nil, err
	}
	comics, err := srv.searchComics(ctx, 10, stemmed)
	if err != nil {
		return nil, fmt.Errorf("error search comics: %w", err)
	}
	return comics, nil
}

func (srv *ComicsService) searchComics(ctx context.Context, n int, keywords []string) ([]domain.Comics, error) {
	select {
	case <-ctx.Done():
		return nil, ErrContextDone
	default:
	}
	idxLastUpdate, err := srv.indexer.GetLastUpdateTime(ctx)
	if err != nil {
		comics, err := srv.searchComicsFromRepository(ctx, n, keywords)
		if err != nil {
			return nil, fmt.Errorf("error search from repository: %w", err)
		}
		return comics, nil
	}
	repoLastUpdate, err := srv.repo.GetLastUpdateTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("error get last update time repo: %w", err)
	}
	if !idxLastUpdate.Equal(repoLastUpdate) {
		comics, err := srv.searchComicsFromRepository(ctx, n, keywords)
		if err != nil {
			return nil, fmt.Errorf("erorr search from repository: %w", err)
		}
		return comics, nil
	}
	indexes, err := srv.indexer.GetNumbersOfNMostRelevantComics(ctx, n, keywords)
	if err != nil {
		comics, err := srv.searchComicsFromRepository(ctx, n, keywords)
		if err != nil {
			return nil, fmt.Errorf("erorr search from repository: %w", err)
		}
		return comics, nil
	}
	comics, err := srv.getComicsByIDs(ctx, indexes)
	if err != nil {
		return nil, fmt.Errorf("error get comics by id: %w", err)
	}
	return comics, nil
}

func (srv *ComicsService) searchComicsFromRepository(ctx context.Context, n int, keywords []string) ([]domain.Comics, error) {
	base := make(map[string]bool, len(keywords))
	for _, keyword := range keywords {
		base[keyword] = true
	}

	comics, err := srv.repo.GetComics(ctx)
	if err != nil {
		return nil, fmt.Errorf("error get comics from repository: %w", err)
	}

	k := make(map[int]int, len(comics))
	for ID, c := range comics {
		for _, keyword := range c.Keywords {
			if base[keyword] {
				k[ID]++
			}
		}
	}
	ans, err := srv.getComicsByIDs(ctx, pair.GetNRelevantFromMap(k, n))
	if err != nil {
		return nil, fmt.Errorf("error get comics by IDs: %w", err)
	}
	return ans, nil
}

func (srv *ComicsService) getComicsByIDs(ctx context.Context, indexes []int) ([]domain.Comics, error) {
	ans := make([]domain.Comics, 0, len(indexes))
	for _, ID := range indexes {
		url, err := srv.repo.GetURLComicsByID(ctx, ID)
		if err != nil {
			return nil, fmt.Errorf("error get comics by id: %w", err)
		}
		ans = append(ans, domain.Comics{ImgURL: url})
	}
	return ans, nil
}

func (srv *ComicsService) UpdateComics(ctx context.Context) (domain.UpdateMeta, error) {
	cntInServer, err := srv.parser.GetCountComicsInServer(ctx)
	if err != nil {
		return domain.UpdateMeta{}, fmt.Errorf("error get count of comics in server: %w", err)
	}

	var parsedComics []domain.Comics
	cnt, err := srv.repo.GetCountComics(ctx)
	if err != nil {
		return domain.UpdateMeta{}, fmt.Errorf("error get count comics in storage: %w", err)
	}
	t, err := srv.repo.GetLastFullCheckTime(ctx)
	if err != nil {
		return domain.UpdateMeta{}, fmt.Errorf("error get last full check time in storage: %w", err)
	}

	if cnt == 0 || srv.checkMonth(t) {
		parsedComics, err = srv.parser.FullParse(ctx, cntInServer)
		if err != nil {
			return domain.UpdateMeta{}, fmt.Errorf("error full parse: %w", err)
		}
		if err = srv.repo.UpdateLastFullCheckTime(ctx, time.Now()); err != nil {
			return domain.UpdateMeta{}, fmt.Errorf("error update last full check time: %w", err)
		}
	} else {
		isNotExists, err := srv.repo.GetIDMissingComics(ctx, cntInServer)
		if err != nil {
			return domain.UpdateMeta{}, fmt.Errorf("error get missing IDs: %w", err)
		}
		parsedComics, err = srv.parser.PartParse(ctx, isNotExists)
		if err != nil {
			return domain.UpdateMeta{}, fmt.Errorf("error part parse: %w", err)
		}
	}

	for _, comics := range parsedComics {
		if err = srv.repo.Add(ctx, comics, comics.ID); err != nil {
			return domain.UpdateMeta{}, fmt.Errorf("error add comics in storage: %w", err)
		}
	}

	if err = srv.updateIndex(ctx, parsedComics); err != nil {
		return domain.UpdateMeta{}, fmt.Errorf("error update index: %w", err)
	}

	var updateTime time.Time
	if len(parsedComics) == 0 {
		updateTime, err = srv.repo.GetLastUpdateTime(ctx)
		if err != nil {
			return domain.UpdateMeta{}, fmt.Errorf("error get last update time from repository: %w", err)
		}
	} else {
		updateTime = time.Now()
	}

	if err = srv.repo.Close(ctx, updateTime); err != nil {
		return domain.UpdateMeta{}, fmt.Errorf("error save comics in storage: %w", err)
	}

	if err = srv.indexer.Save(ctx, updateTime); err != nil {
		return domain.UpdateMeta{}, fmt.Errorf("error save comics in storage: %w", err)
	}

	return domain.UpdateMeta{
		New:   len(parsedComics),
		Total: len(parsedComics) + cnt,
	}, nil
}

func (srv *ComicsService) updateIndex(ctx context.Context, parsedComics []domain.Comics) error {
	idxTime, err := srv.indexer.GetLastUpdateTime(ctx)
	if err != nil {
		return fmt.Errorf("error get last update of index: %w", err)
	}

	repoTime, err := srv.repo.GetLastUpdateTime(ctx)
	if err != nil {
		return fmt.Errorf("error get last update time of repository: %w", err)
	}

	if idxTime != repoTime {
		parsedComics, err = srv.repo.GetComics(ctx)
		if err != nil {
			return fmt.Errorf("error get comics from repository: %w", err)
		}
		if err = srv.indexer.Clear(ctx); err != nil {
			return fmt.Errorf("error clear index: %w", err)
		}
	}

	for _, comics := range parsedComics {
		if err = srv.indexer.UpdateIndex(ctx, comics.ID, comics.Keywords); err != nil {
			return fmt.Errorf("error update index: %w", err)
		}
	}
	return nil
}

func (srv *ComicsService) checkMonth(t time.Time) bool {
	return t.AddDate(0, 1, 0).Before(time.Now())
}
