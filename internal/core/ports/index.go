package ports

import "time"

type Indexer interface {
	GetNumbersOfNMostRelevantComics(n int, keywords []string) ([]int, error)
	UpdateIndex(id int, keywords []string) error
	Save(updateTime time.Time) error
	GetLastUpdateTime() (time.Time, error)
	Clear() error
}
