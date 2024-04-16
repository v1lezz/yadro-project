package xkcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

var ( // errors
	ErrComicsNotFound   = errors.New("comics not found")
	ErrResponseFromXKCD = errors.New("response status is not 200")
	ErrContextDone      = errors.New("contextDone")
)

var ( //url
	urlGetComicsByID = "https://%s/%d/info.0.json"
	urlGetLastComics = "https://%s/info.0.json"
)

type XkcdParse struct {
	URL      string
	Parallel int
}

func NewXkcdParse(url string, parallel int) *XkcdParse {
	return &XkcdParse{
		URL:      url,
		Parallel: parallel,
	}
}

type Semaphore struct {
	ch chan struct{}
}

func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.ch
}

type ResultWithError struct {
	Comics Comics
	Err    error
}

func (xp XkcdParse) FullParse(ctx context.Context, cntInServer int) ([]Comics, error) {
	s := Semaphore{
		ch: make(chan struct{}, xp.Parallel),
	}
	//wg := sync.WaitGroup{}
	outputChan := make(chan ResultWithError, cntInServer)
	//signalChan := make(chan struct{})
	for i := 1; i <= cntInServer; i++ {
		ID := i
		//wg.Add(1)
		go func() {
			//defer wg.Done()
			s.Acquire()
			defer s.Release()
			select {
			case <-ctx.Done():
				outputChan <- ResultWithError{Err: errors.New("context done")}
				return
			default:
			}
			c, err := xp.GetComicsByID(ID)
			if err != nil {
				if errors.Is(err, ErrComicsNotFound) {
					outputChan <- ResultWithError{Err: err}
					return
				}
				log.Println(fmt.Errorf("error get info about comics with id %d: %w", ID, err))
				outputChan <- ResultWithError{Err: fmt.Errorf("error get info about comics with id %d: %w", ID+1, err)}
				return
			}
			//select {
			//case <-signalChan:
			//	return
			//default:
			outputChan <- ResultWithError{
				Comics: c,
				Err:    nil,
			}
			//}
		}()
	}
	ans, err := xp.ReadInArrayFromChan(outputChan, cntInServer)
	close(outputChan)
	if err != nil {
		return nil, err
	}
	return ans, nil
}

func (xp XkcdParse) PartParse(ctx context.Context, isNotExist []int) ([]Comics, error) {
	s := Semaphore{
		ch: make(chan struct{}, xp.Parallel),
	}
	outputChan := make(chan ResultWithError, len(isNotExist))
	for _, ID := range isNotExist {
		currID := ID
		go func() {
			s.Acquire()
			defer s.Release()
			select {
			case <-ctx.Done():
				outputChan <- ResultWithError{Err: ErrContextDone}
				return
			default:
			}
			c, err := xp.GetComicsByID(currID)
			if err != nil {
				if errors.Is(err, ErrComicsNotFound) {
					outputChan <- ResultWithError{Err: err}
					return
				}
				log.Println(fmt.Errorf("error get info about comics with id %d: %w", currID, err))
				outputChan <- ResultWithError{Err: fmt.Errorf("error get info about comics with id %d: %w", currID, err)}
				return
			}
			outputChan <- ResultWithError{
				Comics: c,
				Err:    nil,
			}
		}()
	}
	ans, err := xp.ReadInArrayFromChan(outputChan, len(isNotExist))
	close(outputChan)
	if err != nil {
		return nil, err
	}
	return ans, nil
}

func (xp XkcdParse) ReadInArrayFromChan(outputChan <-chan ResultWithError, n int) ([]Comics, error) {
	ans := make([]Comics, 0, n)
	for i := 0; i < n; i++ {
		res, ok := <-outputChan
		if !ok {
			break
		}
		if res.Err != nil {
			continue
		}
		ans = append(ans, res.Comics)
	}
	return ans, nil
}

func (xp XkcdParse) GetComicsByID(ID int) (Comics, error) {
	resp, err := http.Get(fmt.Sprintf(urlGetComicsByID, xp.URL, ID))
	if err != nil {
		return Comics{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return Comics{}, ErrComicsNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return Comics{}, ErrResponseFromXKCD
	}
	c := Comics{}
	if err = json.NewDecoder(resp.Body).Decode(&c); err != nil {
		return Comics{}, err
	}
	return c, nil
}

func (xp XkcdParse) GetCountComicsInServer(ctx context.Context) (int, error) {
	start := 1
	end := 100
	for ; ; start, end = end, end+100 {
		flag, err := xp.IsNotFoundComics(end, true)
		if err != nil {
			return 0, err
		}
		if flag {
			break
		}
	}
	for start < end {
		middle := (start + end) / 2
		flag, err := xp.IsNotFoundComics(middle, true)
		if err != nil {
			return 0, err
		}
		if flag {
			end = middle - 1
		} else {
			start = middle + 1
		}
	}
	return end - 1, nil
}

func (xp XkcdParse) IsNotFoundComics(ID int, isMain bool) (bool, error) {
	_, err := xp.GetComicsByID(ID)
	if err != nil {
		if errors.Is(err, ErrComicsNotFound) {
			if !isMain {
				return true, nil
			}
			f1, err := xp.IsNotFoundComics(ID-1, false)
			if err != nil {
				return false, err
			}
			f2, err := xp.IsNotFoundComics(ID+1, false)
			if err != nil {
				return false, err
			}
			return f1 && f2, nil
		}
		return false, err
	}
	return false, nil
}

//func (xp XkcdParse) GetCountComicsInServer(ctx context.Context) (int, error) {
//	resp, err := http.Get(fmt.Sprintf(urlGetLastComics, xp.URL))
//	if err != nil {
//		return 0, err
//	}
//	defer resp.Body.Close()
//	if resp.StatusCode != http.StatusOK {
//		return 0, ErrResponseFromXKCD
//	}
//	data := map[string]interface{}{}
//	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
//		return 0, fmt.Errorf("error decode json: %w", err)
//	}
//	return getNumberFromField(data, "num")
//}
//
//func getNumberFromField(data map[string]interface{}, fieldName string) (int, error) {
//	if iNum, ok := data[fieldName]; ok {
//		if num, ok := iNum.(float64); ok {
//			return int(num), nil
//		}
//		log.Println(data["num"])
//		return 0, errors.New(fmt.Sprintf("field \"%s\" is not a number", fieldName))
//	}
//	return 0, errors.New(fmt.Sprintf("field \"%s\" is not exist", fieldName))
//}
