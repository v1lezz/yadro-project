package words

import (
	"fmt"
	"strings"

	"github.com/kljensen/snowball"
	stopchecker "github.com/kljensen/snowball/english"
)

type SnowBallStem struct {
}

func NewSnowBallStem() *SnowBallStem {
	return &SnowBallStem{}
}

func (sbs SnowBallStem) Stem(words []string) ([]string, error) {
	ans := make([]string, 0, len(words))
	was := make(map[string]struct{})
	for _, word := range words {
		stemmed, err := snowball.Stem(word, "english", false)
		if err != nil {
			return nil, fmt.Errorf("error steming: %w", err)
		}

		if _, ok := was[stemmed]; !ok && !stopchecker.IsStopWord(stemmed) && !IsShortStopWord(stemmed) {
			ans = append(ans, stemmed)
			was[stemmed] = struct{}{}
		}
	}
	return ans, nil
}

func IsShortStopWord(line string) bool {
	splitted := strings.Split(line, "'")
	if len(splitted) != 2 {
		return false
	}
	return splitted[1] == "ll" ||
		splitted[1] == "re" ||
		splitted[1] == "m" ||
		splitted[1] == "s" ||
		splitted[1] == "d" ||
		splitted[1] == "ve"
}
