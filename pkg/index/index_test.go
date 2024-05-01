package index

import (
	"testing"
	"yadro-project/internal/config"
)

func BenchmarkFileIndex_GetNumbersOfNMostRelevantComics(b *testing.B) {
	fi, err := NewFileIndex(config.IndexConfig{IndexFile: "../../index.json"})
	if err != nil {
		b.Fatalf("error create index: %s", err.Error())
	}
	for i := 0; i < b.N; i++ {
		if _, err = fi.GetNumbersOfNMostRelevantComics(10, []string{"captcha", "mine"}); err != nil {
			b.Fatalf("error get numbers of n most relevant comics: %s", err.Error())
		}
	}
}
