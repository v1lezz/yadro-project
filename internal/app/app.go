package app

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode"
	"yadro-project/pkg/database"
	"yadro-project/pkg/xkcd"
)

type App struct {
	Parser  Parser
	Stemmer Stemmer
	Repo    Repository
	Indexer Indexer
}

func NewApp(parser Parser, stemmer Stemmer, repo Repository, indexer Indexer) *App {
	return &App{
		Parser:  parser,
		Stemmer: stemmer,
		Repo:    repo,
		Indexer: indexer,
	}
}

type Parser interface {
	GetCountComicsInServer(ctx context.Context) (int, error)
	PartParse(ctx context.Context, isNotExist []int) ([]xkcd.Comics, error)
	FullParse(ctx context.Context, cntInServer int) ([]xkcd.Comics, error)
}

type Stemmer interface {
	Stem([]string) ([]string, error)
}

type Repository interface {
	GetComics() (map[int]database.Comics, error)
	GetCountComics() (int, error)
	GetIDMissingComics(cntInServer int) ([]int, error)
	Add(comics database.Comics, id int) error
	Save(updateTime time.Time) error
	GetLastFullCheckTime() (time.Time, error)
	UpdateLastFullCheckTime(time.Time) error
	GetNumbersOfNMostRelevantComics(n int, keywords []string) ([]int, error)
	GetLastUpdateTime() (time.Time, error)
	GetURLComicsByID(ID int) (string, error)
}

type Indexer interface {
	GetNumbersOfNMostRelevantComics(n int, keywords []string) ([]int, error)
	UpdateIndex(id int, keywords []string) error
	Save(updateTime time.Time) error
	GetLastUpdateTime() (time.Time, error)
	Clear() error
}

func (a *App) Run(stringForSearch string, useIndex bool) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGABRT)
	defer cancel()
	cntInServer, err := a.Parser.GetCountComicsInServer(ctx)
	if err != nil {
		return fmt.Errorf("error get count of comics in server: %w", err)
	}
	//log.Printf("%d comics in server\n", cntInServer)
	var parsedComics []xkcd.Comics
	cnt, err := a.Repo.GetCountComics()
	if err != nil {
		return fmt.Errorf("error get count comics in storage: %w", err)
	}
	t, err := a.Repo.GetLastFullCheckTime()
	if err != nil {
		return fmt.Errorf("error get last full check time in storage: %w", err)
	}
	if cnt == 0 || a.CheckMonth(t) {
		parsedComics, err = a.Parser.FullParse(ctx, cntInServer)
		if err != nil {
			return fmt.Errorf("error full parse: %w", err)
		}
		if err = a.Repo.UpdateLastFullCheckTime(time.Now()); err != nil {
			return fmt.Errorf("error update last full check time: %w", err)
		}
	} else {
		isNotExists, err := a.Repo.GetIDMissingComics(cntInServer)
		if err != nil {
			return fmt.Errorf("error get missing IDs: %w", err)
		}
		parsedComics, err = a.Parser.PartParse(ctx, isNotExists)
		if err != nil {
			return fmt.Errorf("error part parse: %w", err)
		}
	}
	dbComics, err := a.StemSliceComics(parsedComics)
	if err != nil {
		return fmt.Errorf("error stem slice comics: %w", err)
	}
	for ID, comics := range dbComics {
		if err = a.Repo.Add(comics, ID); err != nil {
			return fmt.Errorf("error add comics in storage: %w", err)
		}

	}
	if err = a.UpdateIndex(dbComics); err != nil {
		return err
	}
	var updateTime time.Time
	if len(dbComics) == 0 {
		updateTime, err = a.Repo.GetLastUpdateTime()
		if err != nil {
			return err
		}
	} else {
		updateTime = time.Now()
	}

	if err = a.Repo.Save(updateTime); err != nil {
		return fmt.Errorf("error save comics in storage: %w", err)
	}
	if err = a.Indexer.Save(updateTime); err != nil {
		return fmt.Errorf("error save index in storage: %w", err)
	}

	// если поступил сигнал, мы только сохраняем полученные результаты
	// но не выводим, поскольку мы могли получить не все комиксы
	select {
	case <-ctx.Done():
		fmt.Println("interrupted")
		return nil
	default:
	}
	if stringForSearch == "" {
		return nil
	}
	IDs, err := a.GetNumbersOfNRelevantComics(10, stringForSearch, useIndex)
	for _, ID := range IDs {
		url, err := a.Repo.GetURLComicsByID(ID)
		if err != nil {
			continue
		}
		fmt.Println(url)
	}
	return nil
}

func (a *App) UpdateIndex(dbComics map[int]database.Comics) error {
	idxTime, err := a.Indexer.GetLastUpdateTime()
	if err != nil {
		return fmt.Errorf("error get last update time of index: %w", err)
	}
	repoTime, err := a.Repo.GetLastUpdateTime()
	if err != nil {
		return fmt.Errorf("error get last update time of repository: %w", err)
	}
	if idxTime != repoTime {
		dbComics, err = a.Repo.GetComics()
		if err = a.Indexer.Clear(); err != nil {
			return fmt.Errorf("error clear index: %w", err)
		}
	}
	for ID, comics := range dbComics {
		if err = a.Indexer.UpdateIndex(ID, comics.Keywords); err != nil {
			return fmt.Errorf("error update index: %w", err)
		}
	}
	return nil
}

func (a *App) StemSliceComics(parsedComics []xkcd.Comics) (map[int]database.Comics, error) {
	ans := make(map[int]database.Comics, len(parsedComics))
	for _, comics := range parsedComics {
		cAns, err := a.StemComics(comics)
		if err == nil {
			ans[comics.ID] = cAns
		}
	}
	return ans, nil
}

func (a *App) StemComics(comics xkcd.Comics) (database.Comics, error) {
	cAns := database.Comics{
		ImgURL: comics.ImgURL,
	}
	keywords, err := a.Stemmer.Stem(comics.GetWordsFromTranscriptAndAlt())
	if err != nil {
		return database.Comics{}, fmt.Errorf("error stem comics with id %d:%w", comics.ID, err)
	}
	cAns.Keywords = keywords
	return cAns, nil
}

func (a App) CheckMonth(t time.Time) bool {
	return t.AddDate(0, 1, 0).Before(time.Now())
}

func (a *App) GetNumbersOfNRelevantComics(n int, stringForSearch string, useIndex bool) ([]int, error) {
	keywords := strings.FieldsFunc(stringForSearch, func(r rune) bool {
		return !unicode.IsLetter(r) && r != '\''
	})
	keywords, err := a.Stemmer.Stem(keywords)
	if err != nil {
		return nil, fmt.Errorf("error stem string for search: %w", err)
	}
	if useIndex {
		ans, err := a.Indexer.GetNumbersOfNMostRelevantComics(n, keywords)
		if err == nil {
			return ans, nil
		}
		log.Printf("error get relevant comics from index: %s", err.Error())
	}
	ans, err := a.Repo.GetNumbersOfNMostRelevantComics(n, keywords)
	if err != nil {
		return nil, err
	}
	return ans, nil
}
