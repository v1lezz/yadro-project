package index

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"yadro-project/internal/adapters/repository"
	"yadro-project/internal/config"
	"yadro-project/pkg/pair"
)

type Data struct {
	LastUpdate time.Time
	Index      map[string][]int
}

type FileIndex struct {
	IndexFilePath string
	Data          Data
	wasChanged    bool
}

func NewFileIndex(cfg config.IndexConfig) (*FileIndex, error) {
	idx := &FileIndex{
		IndexFilePath: cfg.IndexFile,
		Data: Data{
			Index: make(map[string][]int),
		},
		wasChanged: false,
	}
	flag, err := repository.FileIsExist(cfg.IndexFile)
	if err != nil {
		return nil, err
	}
	if !flag {
		return idx, nil
	}
	file, err := os.Open(cfg.IndexFile)
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("error open file \"%s\": %w", cfg.IndexFile, err)
	}
	if err = json.NewDecoder(file).Decode(&idx.Data); err != nil {
		return nil, fmt.Errorf("error decode json from \"%s\": %w", cfg.IndexFile, err)
	}
	return idx, nil
}

func (fi *FileIndex) GetNumbersOfNMostRelevantComics(n int, keywords []string) ([]int, error) {
	cnt := make(map[int]int)
	for _, keyword := range keywords {
		for _, number := range fi.Data.Index[keyword] {
			cnt[number]++
		}
	}
	return pair.GetNRelevantFromMap(cnt, n), nil
}

func (fi *FileIndex) UpdateIndex(id int, keywords []string) error {
	for _, keyword := range keywords {
		t := fi.Data.Index[keyword]
		t = append(t, id)
		fi.Data.Index[keyword] = t
	}
	fi.wasChanged = true
	return nil
}

func (fi *FileIndex) Save(updateTime time.Time) error {
	if !fi.wasChanged {
		return nil
	}
	fi.Data.LastUpdate = updateTime
	file, err := os.OpenFile(fi.IndexFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	return json.NewEncoder(file).Encode(&fi.Data)
}

func (fi *FileIndex) GetLastUpdateTime() (time.Time, error) {
	return fi.Data.LastUpdate, nil
}

func (fi *FileIndex) Clear() error {
	fi.Data.Index = make(map[string][]int)
	return nil
}
