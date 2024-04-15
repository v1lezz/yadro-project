package app

import (
	"fmt"
	"io"
	"strconv"
	"time"
	"yadro-project/pkg/database"
	"yadro-project/pkg/xkcd"
)

type App struct {
	Parser        Parser
	FlagParser    FlagParser
	ComicsStemmer Stemmer
	Repo          Repository
	Writer        io.WriteCloser
}

func NewApp(parser Parser, stemmer Stemmer, repo Repository, flagParser FlagParser, writer io.WriteCloser) *App {
	return &App{
		Parser:        parser,
		Writer:        writer,
		ComicsStemmer: stemmer,
		FlagParser:    flagParser,
		Repo:          repo,
	}
}

type Parser interface {
	GetCountComicsInServer() (int, error)
	PartParse([]uint, int, int) ([]xkcd.Comics, error)
	FullParse(cntInServer, start, end int) ([]xkcd.Comics, error)
}

type Stemmer interface {
	SliceComicsStem(map[string]database.Comics, []xkcd.Comics) error
}

type Repository interface {
	GetComics() (map[string]database.Comics, time.Time, error)
	SaveComics(map[string]database.Comics, time.Time) error
}

type FlagParser interface {
	ParseFlagOandN(int) (bool, int, error)
}

func (a *App) Run() error {
	cntInServer, err := a.Parser.GetCountComicsInServer()
	if err != nil {
		return fmt.Errorf("error get count of comics in server: %w", err)
	}
	o, n, err := a.FlagParser.ParseFlagOandN(cntInServer)
	if err != nil {
		return err
	}
	dbComics, t, err := a.Repo.GetComics()
	if err != nil {
		return fmt.Errorf("error get comics from database: %w", err)
	}
	var parsedComics []xkcd.Comics
	var tAns time.Time
	if len(dbComics) == 0 || a.CheckMonth(t) {
		parsedComics, err = a.Parser.FullParse(cntInServer, 1, n)
		tAns = time.Now()
	} else {
		isNotExists := database.CheckComics(dbComics, n)
		parsedComics, err = a.Parser.PartParse(isNotExists, cntInServer, n)
		tAns = t
	}
	if err = a.ComicsStemmer.SliceComicsStem(dbComics, parsedComics); err != nil {
		return err
	}
	if err = a.Repo.SaveComics(dbComics, tAns); err != nil {
		return err
	}
	if o {
		for i, cntOutput := 0, 0; cntOutput < min(n, len(dbComics)); i++ {
			if v, ok := dbComics[strconv.Itoa(i+1)]; ok {
				_, err = fmt.Fprintf(a.Writer, "%s\n\n", v.String(i+1))
				if err != nil {
					return err
				}
				cntOutput++
			}

		}
	}
	return nil
}

func (a App) CheckMonth(t time.Time) bool {
	return t.AddDate(0, 1, 0).Before(time.Now())
}

func (a *App) Close() {
	a.Writer.Close()
}
