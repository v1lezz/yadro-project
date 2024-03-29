package main

import (
	"fmt"
	"github.com/kljensen/snowball"
	stop_checker "github.com/kljensen/snowball/english"
	"strings"
)

type SnowBallStem struct {
}

func NewSnowBallStem() *SnowBallStem {
	return &SnowBallStem{}
}

func (sbs SnowBallStem) Stem(words []string) (string, error) {
	ans := make([]string, 0, len(words))
	was := make(map[string]struct{})
	for _, word := range words {
		stemmed, err := snowball.Stem(word, "english", false)
		if err != nil {
			return "", fmt.Errorf("error steming: %w", err)
		}
		//log.Printf("stteming word:\"%s\", stemmed:\"%s\"\n", word, stemmed)

		if _, ok := was[stemmed]; !ok && !stop_checker.IsStopWord(stemmed) && !strings.Contains(stemmed, "'") {
			ans = append(ans, stemmed)
			was[stemmed] = struct{}{}
		}
	}
	return strings.Join(ans, " "), nil
}
