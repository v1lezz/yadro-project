package xkcd

import (
	"strings"
	"unicode"
)

type Comics struct {
	ID         int    `json:"num"`
	ImgURL     string `json:"img"`
	Transcript string `json:"transcript"`
	Alt        string `json:"title"`
}

func (c Comics) GetWordsFromTranscriptAndAlt() []string {
	splitted := strings.FieldsFunc(c.Transcript, func(r rune) bool {
		return !unicode.IsLetter(r) && r != '\''
	})
	flag := false
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
		splitted = append(splitted, strings.FieldsFunc(c.Alt, func(r rune) bool {
			return !unicode.IsLetter(r) && r != '\''
		})...)
	}
	return splitted
}
