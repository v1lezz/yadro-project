package xkcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"yadro-project/internal/core/domain"
	"yadro-project/internal/core/ports"
)

var ( // errors
	ErrComicsNotFound   = errors.New("comics not found")
	ErrResponseFromXKCD = errors.New("response status is not 200")
	ErrContextDone      = errors.New("contextDone")
)

var ( //url
	urlGetComicsByID = "https://%s/%d/info.0.json"
)

type XkcdParse struct {
	URL      string
	Parallel int
	Stemmer  ports.Stemmer
}

func NewXkcdParse(url string, parallel int, stemmer ports.Stemmer) *XkcdParse {
	return &XkcdParse{
		URL:      url,
		Parallel: parallel,
		Stemmer:  stemmer,
	}
}

func (xp *XkcdParse) stemSliceComics(parsedComics []Comics) ([]domain.Comics, error) {
	ans := make([]domain.Comics, 0, len(parsedComics))
	for _, comics := range parsedComics {
		cAns, err := xp.stemComics(comics)
		if err == nil {
			ans = append(ans, cAns)
		}
	}
	return ans, nil
}

func (xp *XkcdParse) stemComics(comics Comics) (domain.Comics, error) {
	cAns := domain.Comics{
		ID:     comics.ID,
		ImgURL: comics.ImgURL,
	}
	keywords, err := xp.Stemmer.Stem(comics.GetWordsFromTranscriptAndAlt())
	if err != nil {
		return domain.Comics{}, fmt.Errorf("error stem comics with id %d:%w", comics.ID, err)
	}
	cAns.Keywords = keywords
	return cAns, nil
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

func (xp *XkcdParse) FullParse(ctx context.Context, cntInServer int) ([]domain.Comics, error) {
	s := Semaphore{
		ch: make(chan struct{}, xp.Parallel),
	}
	outputChan := make(chan ResultWithError, cntInServer)
	for i := 1; i <= cntInServer; i++ {
		ID := i
		go func() {
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
			outputChan <- ResultWithError{
				Comics: c,
				Err:    nil,
			}

		}()
	}
	ans, err := xp.ReadInArrayFromChan(outputChan, cntInServer)
	close(outputChan)
	if err != nil {
		return nil, err
	}
	return xp.stemSliceComics(ans)
}

func (xp *XkcdParse) PartParse(ctx context.Context, isNotExist []int) ([]domain.Comics, error) {
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
	return xp.stemSliceComics(ans)
}

func (xp *XkcdParse) ReadInArrayFromChan(outputChan <-chan ResultWithError, n int) ([]Comics, error) {
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

func (xp *XkcdParse) GetComicsByID(ID int) (Comics, error) {
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

func (xp *XkcdParse) GetCountComicsInServer(ctx context.Context) (int, error) {
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

func (xp *XkcdParse) IsNotFoundComics(ID int, isMain bool) (bool, error) {
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
