package app

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"
	"yadro-project/pkg/database"
	"yadro-project/pkg/xkcd"
)

type App struct {
	Parser  Parser
	Stemmer Stemmer
	Repo    Repository
}

func NewApp(parser Parser, stemmer Stemmer, repo Repository) *App {
	return &App{
		Parser:  parser,
		Stemmer: stemmer,
		Repo:    repo,
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
	Save() error
	GetLastFullCheckTime() (time.Time, error)
	UpdateLastFullCheckTime(time.Time) error
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGABRT)
	defer cancel()
	cntInServer, err := a.Parser.GetCountComicsInServer(ctx)
	if err != nil {
		return fmt.Errorf("error get count of comics in server: %w", err)
	}
	log.Printf("%d comics in server\n", cntInServer)
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
	if err = a.Repo.Save(); err != nil {
		return fmt.Errorf("error save comics in storage: %w", err)
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
