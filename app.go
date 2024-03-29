package main

import (
	"bufio"
	"fmt"
)

type App struct {
	Parser  Parser
	Stemmer Stemmer
	Writer  *bufio.Writer
}

func NewApp(parser Parser, stemmer Stemmer, writer *bufio.Writer) *App {
	return &App{
		Parser:  parser,
		Writer:  writer,
		Stemmer: stemmer,
	}
}

type Parser interface {
	Parse() ([]string, error)
}

type Stemmer interface {
	Stem([]string) (string, error)
}

func (a App) Run() error {
	parsed, err := a.Parser.Parse()
	if err != nil {
		return fmt.Errorf("error parsing value: %w", err)
	}
	ans, err := a.Stemmer.Stem(parsed)
	if err != nil {
		return fmt.Errorf("error stem: %w", err)
	}
	if _, err = fmt.Fprintln(a.Writer, ans); err != nil {
		return fmt.Errorf("error writing, %w", err)
	}
	return nil
}

func (a App) Close() {
	a.Writer.Flush()
}
