package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
	"yadro-project/internal/core/domain"
)

type data struct {
	LastFullCheck time.Time             `json:"last_full_check"`
	LastUpdate    time.Time             `json:"last_update"`
	Comics        map[int]domain.Comics `json:"comics"`
}

type JsonDB struct {
	JsonFilePath string
	Data         data
	SliceComics  []domain.Comics
	wasChanged   bool
}

func NewJsonDB(filePath string) (*JsonDB, error) {
	flag, err := FileIsExist(filePath)
	if err != nil {
		return nil, err
	}
	if !flag {
		return &JsonDB{
			JsonFilePath: filePath,
			Data: data{
				LastUpdate:    time.Time{},
				LastFullCheck: time.Time{},
				Comics:        make(map[int]domain.Comics),
			},
			SliceComics: make([]domain.Comics, 0),
			wasChanged:  false,
		}, nil
	}
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("error open file \"%s\": %w", filePath, err)
	}
	jdb := &JsonDB{
		JsonFilePath: filePath,
		Data: data{
			Comics: make(map[int]domain.Comics),
		},
	}
	if err = json.NewDecoder(file).Decode(&jdb.Data); err != nil {
		return nil, fmt.Errorf("error decode json from \"%s\": %w", filePath, err)
	}
	sliceComics := make([]domain.Comics, 0, len(jdb.Data.Comics))
	for ID, comics := range jdb.Data.Comics {
		comics.ID = ID
		jdb.Data.Comics[ID] = comics
		sliceComics = append(sliceComics, comics)
	}
	return jdb, nil
}

func (db *JsonDB) GetComics(ctx context.Context) ([]domain.Comics, error) {
	return db.SliceComics, nil
}

func (db *JsonDB) Add(ctx context.Context, comics domain.Comics, id int) error {
	if _, ok := db.Data.Comics[id]; ok {
		return errors.New(fmt.Sprintf("Comics with id %d already exist", id))
	}
	db.Data.Comics[id] = comics
	db.SliceComics = append(db.SliceComics, comics)
	db.wasChanged = true
	return nil
}

func (db *JsonDB) Close(ctx context.Context, updateTime time.Time) error {
	if !db.wasChanged {
		return nil
	}
	db.Data.LastUpdate = updateTime
	file, err := os.OpenFile(db.JsonFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return err
	}
	return json.NewEncoder(file).Encode(&db.Data)
}

//func (db *JsonDB) GetNumbersOfNMostRelevantComics(n int, keywords []string) ([]int, error) {
//	base := make(map[string]bool, len(keywords))
//	for _, keyword := range keywords {
//		base[keyword] = true
//	}
//	k := make(map[int]int, len(db.Data.Comics))
//	for ID, comics := range db.Data.Comics {
//		cnt := 0
//		for _, keyword := range comics.Keywords {
//			if base[keyword] {
//				cnt++
//			}
//		}
//		if cnt != 0 {
//			k[ID] = cnt
//		}
//	}
//	return pair.GetNRelevantFromMap(k, n), nil
//}

func (db *JsonDB) UpdateLastFullCheckTime(ctx context.Context, t time.Time) error {
	db.Data.LastFullCheck = t
	return nil
}

func (db *JsonDB) GetLastFullCheckTime(ctx context.Context) (time.Time, error) {
	return db.Data.LastFullCheck, nil
}

func (db *JsonDB) GetCountComics(ctx context.Context) (int, error) {
	return len(db.Data.Comics), nil
}

func (db *JsonDB) GetIDMissingComics(ctx context.Context, cntInServer int) ([]int, error) {
	ans := make([]int, 0, cntInServer-len(db.Data.Comics))
	for i := 1; i <= cntInServer; i++ {
		if _, ok := db.Data.Comics[i]; !ok {
			ans = append(ans, i)
		}
	}
	return ans, nil
}

func (db *JsonDB) GetLastUpdateTime(ctx context.Context) (time.Time, error) {
	return db.Data.LastUpdate, nil
}

func (db *JsonDB) GetURLComicsByID(ctx context.Context, ID int) (string, error) {
	if val, ok := db.Data.Comics[ID]; ok {
		return val.ImgURL, nil
	} else {
		return "", errors.New("comics not found")
	}
}
