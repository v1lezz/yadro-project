package flag_parser

import (
	"flag"
)

type FlagParser struct {
}

func NewFlagParser() *FlagParser {
	return &FlagParser{}
}

func (fp FlagParser) ParseFlagOandN(defaultN int) (bool, int, error) {
	var o bool
	var n int
	flag.BoolVar(&o, "o", false, "need output")
	flag.IntVar(&n, "n", defaultN, "count of output")

	flag.Parse()
	return o, n, nil
}
