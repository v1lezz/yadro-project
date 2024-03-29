package main

import (
	"flag"
	"strings"
)

type FlagParse struct {
}

func NewFlagParse() *FlagParse {
	return &FlagParse{}
}

func (fg FlagParse) Parse() ([]string, error) {
	first := flag.String("s", "", "parsed string")
	flag.Parse()
	ans := strings.Split(*first, " ")
	//log.Printf("flag.String(): \"%s\" flag.Args():\"%s\"\n", *first, strings.Join(flag.Args(), " "))
	ans = append(ans, flag.Args()...)
	//log.Printf("parsed string:\"%s\"\n", strings.Join(ans, " "))
	return ans, nil
}
