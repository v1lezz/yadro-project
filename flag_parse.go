package main

import (
	"flag"
	"strings"
	"unicode"
)

type FlagParse struct {
}

func NewFlagParse() *FlagParse {
	return &FlagParse{}
}

func (fg FlagParse) Parse() ([]string, error) {
	var ans string
	flag.StringVar(&ans, "s", "", "parsed string")
	flag.Parse()
	return strings.FieldsFunc(ans, func(r rune) bool {
		return !unicode.IsLetter(r) && r != '\''
	}), nil
}
