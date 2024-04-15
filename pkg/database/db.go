package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type JsonDB struct {
	JsonFilePath string
}

func NewJsonDB(filePath string) *JsonDB {
	return &JsonDB{
		JsonFilePath: filePath,
	}
}

func (db JsonDB) GetComics() (map[string]Comics, time.Time, error) {
	flag, err := db.FileIsExist()
	if err != nil {
		return nil, time.Time{}, err
	}
	if !flag {
		return map[string]Comics{}, time.Time{}, nil
	}
	file, err := os.Open(db.JsonFilePath)
	defer file.Close()
	if err != nil {
		return nil, time.Time{}, err
	}
	data := map[string]interface{}{}

	if err = json.NewDecoder(file).Decode(&data); err != nil {
		return nil, time.Time{}, err
	}
	sTime := data["last_full_check"].(string)
	var t time.Time
	if err = t.UnmarshalText([]byte(sTime)); err != nil {
		log.Println(fmt.Errorf("error check time of last full check: %w", err))
	}
	delete(data, "last_full_check")
	ans := make(map[string]Comics)
	wg := sync.WaitGroup{}
	m := sync.Mutex{}
	for ID, comics := range data {
		ID := ID
		comics := comics
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, ok := comics.(map[string]interface{})
			if !ok {
				return
			}
			if v, err := NewComics(c); err == nil {
				m.Lock()
				ans[ID] = v
				m.Unlock()
			}
		}()

	}
	wg.Wait()
	return ans, t, nil
}

func (db JsonDB) SaveComics(data map[string]Comics, t time.Time) error {
	ans := make(map[string]interface{})
	for ID, comics := range data {
		ans[ID] = comics
	}
	ans["last_full_check"] = t
	file, err := os.OpenFile(db.JsonFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return err
	}
	if err = json.NewEncoder(file).Encode(ans); err != nil {
		return err
	}
	return nil
}

func (db JsonDB) FileIsExist() (bool, error) {
	_, err := os.Stat(db.JsonFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
