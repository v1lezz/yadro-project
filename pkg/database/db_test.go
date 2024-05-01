package database

import (
	"testing"
)

func BenchmarkFileIndex_GetNumbersOfNMostRelevantComics(b *testing.B) {
	jdb, err := NewJsonDB("../../database.json")
	if err != nil {
		b.Fatalf("error create database: %s", err.Error())
	}
	for i := 0; i < b.N; i++ {
		if _, err = jdb.GetNumbersOfNMostRelevantComics(10, []string{"captcha", "mine"}); err != nil {
			b.Fatalf("error get numbers of n most relevant comics: %s", err.Error())
		}
	}
}
