package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
	"yadro-project/pkg/pair"
)

type data struct {
	LastFullCheck time.Time      `json:"last_full_check"`
	LastUpdate    time.Time      `json:"last_update"`
	Comics        map[int]Comics `json:"comics"`
}

type JsonDB struct {
	JsonFilePath string
	Data         data
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
				Comics:        make(map[int]Comics),
			},
			wasChanged: false,
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
			Comics: make(map[int]Comics),
		},
	}
	if err = json.NewDecoder(file).Decode(&jdb.Data); err != nil {
		return nil, fmt.Errorf("error decode json from \"%s\": %w", filePath, err)
	}

	return jdb, nil
}

func (db *JsonDB) GetComics() (map[int]Comics, error) {
	return db.Data.Comics, nil
}

func (db *JsonDB) Add(comics Comics, id int) error {
	if _, ok := db.Data.Comics[id]; ok {
		return errors.New(fmt.Sprintf("Comics with id %d already exitst", id))
	}
	db.Data.Comics[id] = comics
	db.wasChanged = true
	return nil
}

func (db *JsonDB) Save(updateTime time.Time) error {
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

func (db *JsonDB) GetNumbersOfNMostRelevantComics(n int, keywords []string) ([]int, error) {
	base := make(map[string]bool, len(keywords))
	for _, keyword := range keywords {
		base[keyword] = true
	}
	k := make(map[int]int, len(db.Data.Comics))
	for ID, comics := range db.Data.Comics {
		cnt := 0
		for _, keyword := range comics.Keywords {
			if base[keyword] {
				cnt++
			}
		}
		if cnt != 0 {
			k[ID] = cnt
		}
	}
	return pair.GetNRelevantFromMap(k, n), nil
}

func (db *JsonDB) UpdateLastFullCheckTime(t time.Time) error {
	db.Data.LastFullCheck = t
	return nil
}

func (db *JsonDB) GetLastFullCheckTime() (time.Time, error) {
	return db.Data.LastFullCheck, nil
}

func (db *JsonDB) GetCountComics() (int, error) {
	return len(db.Data.Comics), nil
}

func (db *JsonDB) GetIDMissingComics(cntInServer int) ([]int, error) {
	ans := make([]int, 0, cntInServer-len(db.Data.Comics))
	for i := 1; i <= cntInServer; i++ {
		if _, ok := db.Data.Comics[i]; !ok {
			ans = append(ans, i)
		}
	}
	return ans, nil
}

func (db *JsonDB) GetLastUpdateTime() (time.Time, error) {
	return db.Data.LastUpdate, nil
}

func (db *JsonDB) GetURLComicsByID(ID int) (string, error) {
	if val, ok := db.Data.Comics[ID]; ok {
		return val.ImgURL, nil
	} else {
		return "", errors.New("comics not found")
	}
}

func FileIsExist(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
