package xkcd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
)

var ( // errors
	ErrComicsNotFound   = errors.New("comics not found")
	ErrResponseFromXKCD = errors.New("response status is not 200")
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

func (xp XkcdParse) FullParse(cntInServer, start, end int) ([]Comics, error) {
	n := end - start + 1
	s := Semaphore{
		ch: make(chan struct{}, xp.Parallel),
	}
	wg := sync.WaitGroup{}
	outputChan := make(chan ResultWithError, n)
	signalChan := make(chan struct{})
	for i := start; i <= min(end, cntInServer); i++ {
		ID := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Acquire()
			defer s.Release()
			c, err := xp.GetComicsByID(uint(ID))
			if err != nil {
				if errors.Is(err, ErrComicsNotFound) {
					//if lastID < cntInServer {
					//	lastID++
					//}
					outputChan <- ResultWithError{
						Comics: Comics{},
						Err:    err,
					}
				} else {
					outputChan <- ResultWithError{
						Comics: Comics{},
						Err:    fmt.Errorf("error get info about comics with id %d: %w", ID+1, err),
					}
				}
				return
			}
			select {
			case <-signalChan:
				return
			default:
				outputChan <- ResultWithError{
					Comics: c,
					Err:    nil,
				}
			}
		}()
	}
	ans, err := xp.ReadInArrayFromChan(&wg, outputChan, signalChan, n)
	if err != nil {
		return nil, err
	}
	if len(ans) < n {
		add, err := xp.FullParse(cntInServer, end+1, end+n-len(ans))
		if err != nil {
			return nil, err
		}
		ans = append(ans, add...)
	}
	return ans, nil
}

func (xp XkcdParse) PartParse(isNotExist []uint, cntInServer, n int) ([]Comics, error) {
	s := Semaphore{
		ch: make(chan struct{}, xp.Parallel),
	}
	wg := sync.WaitGroup{}
	outputChan := make(chan ResultWithError, n)
	signalChan := make(chan struct{})
	for _, ID := range isNotExist {
		currID := ID
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Acquire()
			defer s.Release()
			c, err := xp.GetComicsByID(currID)
			if err != nil {
				if errors.Is(err, ErrComicsNotFound) {
					outputChan <- ResultWithError{
						Comics: Comics{},
						Err:    err,
					}
				} else {
					outputChan <- ResultWithError{
						Comics: Comics{},
						Err:    fmt.Errorf("error get info about comics with id %d: %w", currID, err),
					}
				}
				return
			}
			outputChan <- ResultWithError{
				Comics: c,
				Err:    nil,
			}
		}()
	}
	ans, err := xp.ReadInArrayFromChan(&wg, outputChan, signalChan, len(isNotExist))
	if err != nil {
		return nil, err
	}
	if len(ans) < n {
		add, err := xp.FullParse(cntInServer, n+1, n+len(isNotExist)-len(ans))
		if err != nil {
			return nil, err
		}
		ans = append(ans, add...)
	}
	return ans, nil
}

func (xp XkcdParse) ReadInArrayFromChan(wg *sync.WaitGroup, outputChan <-chan ResultWithError, signalChan chan<- struct{}, n int) ([]Comics, error) {
	ans := make([]Comics, 0, n)
	waitingChan := make(chan struct{})
	go func() {
		wg.Wait()
		waitingChan <- struct{}{}
	}()
	for i := 0; len(ans) < n; i++ {
		select {
		case res, ok := <-outputChan:
			if !ok {
				break
			}
			if res.Err != nil {
				if errors.Is(res.Err, ErrComicsNotFound) {
					continue
				}
				close(signalChan)
				return nil, res.Err
			}
			ans = append(ans, res.Comics)
		case <-waitingChan:
			break
		}
	}
	return ans, nil
}

func (xp XkcdParse) GetComicsByID(ID uint) (Comics, error) {
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

func (xp XkcdParse) GetCountComicsInServer() (int, error) {
	resp, err := http.Get(fmt.Sprintf(urlGetLastComics, xp.URL))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, ErrResponseFromXKCD
	}
	data := map[string]interface{}{}
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("error decode json: %w", err)
	}
	return getNumberFromField(data, "num")
}

func getNumberFromField(data map[string]interface{}, fieldName string) (int, error) {
	if iNum, ok := data[fieldName]; ok {
		if num, ok := iNum.(float64); ok {
			return int(num), nil
		}
		log.Println(data["num"])
		return 0, errors.New(fmt.Sprintf("field \"%s\" is not a number", fieldName))
	}
	return 0, errors.New(fmt.Sprintf("field \"%s\" is not exist", fieldName))
}
