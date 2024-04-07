package words

import (
	"fmt"
	"strings"
	"unicode"
	"yadro-project/pkg/database"
	"yadro-project/pkg/xkcd"

	"github.com/kljensen/snowball"
	stopchecker "github.com/kljensen/snowball/english"
)

type SnowBallStem struct {
}

func NewSnowBallStem() *SnowBallStem {
	return &SnowBallStem{}
}

func (sbs SnowBallStem) SliceComicsStem(data map[string]database.Comics, comics []xkcd.Comics) error {
	//ans := make(map[string]database.Comics, len(comics))
	for _, c := range comics {
		cAns, err := sbs.ComicsStem(c)
		if err != nil {
			return fmt.Errorf("error stemming comics with id %d: %w", c.ID, err)
		}
		data[fmt.Sprintf("%d", c.ID)] = cAns
	}
	return nil
}

//func (sbs SnowBallStem) SliceComicsStem(comics []xkcd.Comics) (map[string]database.Comics, error) {
//	ans := make(map[string]database.Comics, len(comics))
//	for _, c := range comics {
//		cAns, err := sbs.ComicsStem(c)
//		if err != nil {
//			return nil, fmt.Errorf("error stemming comics with id %d: %w", c.ID, err)
//		}
//		ans[fmt.Sprintf("%d", c.ID)] = cAns
//	}
//	return ans, nil
//}

func (sbs SnowBallStem) ComicsStem(comics xkcd.Comics) (database.Comics, error) {
	cAns := database.Comics{
		ImgURL: comics.ImgURL,
	}
	keywords, err := sbs.WordsStem(sbs.GetWordsFromTranscriptAndAlt(comics.Transcript, comics.Alt))
	if err != nil {
		return database.Comics{}, err
	}
	cAns.Keywords = keywords
	return cAns, nil
}

func (sbs SnowBallStem) GetWordsFromTranscriptAndAlt(transcript, alt string) []string {
	splitted := strings.FieldsFunc(transcript, func(r rune) bool {
		return !unicode.IsLetter(r) && r != '\''
	})
	flag := false
	//fmt.Println(splitted)
	for i := len(splitted) - 1; i >= 0; i-- {
		if splitted[i] == "alt" || splitted[i] == "title" {
			flag = true
			if i < len(splitted)-1 && splitted[i+1] == "text" {
				splitted = append(splitted[:i], splitted[i+2:]...)
			} else {
				splitted = append(splitted[:i], splitted[i+1:]...)
			}
			break
		}
	}
	if !flag {
		splitted = append(splitted, strings.FieldsFunc(alt, func(r rune) bool {
			return !unicode.IsLetter(r) && r != '\''
		})...)
	}
	return splitted
}

//func (sbs SnowBallStem) HaveAltInTranscript(transcript string) bool {
//	splitted := strings.Split(transcript, "{{")
//}

func (sbs SnowBallStem) WordsStem(words []string) ([]string, error) {
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
