package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type data struct {
	LastFullCheck time.Time      `json:"last_full_check"`
	Comics        map[int]Comics `json:"comics"`
}

type JsonDB struct {
	JsonFilePath string
	Data         data
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
				LastFullCheck: time.Time{},
				Comics:        make(map[int]Comics),
			},
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
	//flag, err := db.FileIsExist()
	//if err != nil {
	//	return nil, time.Time{}, err
	//}
	//if !flag {
	//	return map[string]Comics{}, time.Time{}, nil
	//}
	//file, err := os.Open(db.JsonFilePath)
	//defer file.Close()
	//if err != nil {
	//	return nil, time.Time{}, err
	//}
	//data := map[string]interface{}{}
	//
	//if err = json.NewDecoder(file).Decode(&data); err != nil {
	//	return nil, time.Time{}, err
	//}
	//sTime := data["last_full_check"].(string)
	//var t time.Time
	//if err = t.UnmarshalText([]byte(sTime)); err != nil {
	//	log.Println(fmt.Errorf("error check time of last full check: %w", err))
	//}
	//delete(data, "last_full_check")
	//ans := make(map[string]Comics)
	//wg := sync.WaitGroup{}
	//m := sync.Mutex{}
	//for ID, comics := range data {
	//	ID := ID
	//	comics := comics
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		c, ok := comics.(map[string]interface{})
	//		if !ok {
	//			return
	//		}
	//		if v, err := NewComics(c); err == nil {
	//			m.Lock()
	//			ans[ID] = v
	//			m.Unlock()
	//		}
	//	}()
	//
	//}
	//wg.Wait()
	//return ans, t, nil
}

func (db *JsonDB) Add(comics Comics, id int) error {
	if _, ok := db.Data.Comics[id]; ok {
		return errors.New(fmt.Sprintf("Comics with id %d already exitst", id))
	}
	db.Data.Comics[id] = comics
	return nil
}

func (db *JsonDB) Save() error {
	file, err := os.OpenFile(db.JsonFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return err
	}
	return json.NewEncoder(file).Encode(&db.Data)
	//ans := make(map[string]interface{})
	//for ID, comics := range data {
	//	ans[ID] = comics
	//}
	//ans["last_full_check"] = t
	//file, err := os.OpenFile(db.JsonFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	//defer file.Close()
	//if err != nil {
	//	return err
	//}
	//if err = json.NewEncoder(file).Encode(ans); err != nil {
	//	return err
	//}
	//return nil
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
