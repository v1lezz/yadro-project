package xkcd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

var ( // errors
	ErrComicsNotFound   = errors.New("comics not found")
	ErrResponseFromXKCD = errors.New("response status is not 200")
)

type XkcdParse struct {
	URL string
}

func NewXkcdParse(url string) *XkcdParse {
	return &XkcdParse{URL: url}
}

func (xp XkcdParse) FullParse(cntInServer, n int) ([]Comics, error) {
	g := &errgroup.Group{}
	ans := make([]Comics, n)
	cntGetted := atomic.Int32{}
	ch := make(chan Comics)
	wg := sync.WaitGroup{}
	wg.Add(1)
	ReturnArrayAfterWriteFromChan(&wg, &ans, ch, 0)
	for i := 0; i < n; i++ {
		i := i
		//fmt.Println(i)
		g.Go(func() error {
			c, err := xp.GetComicsByID(uint(i + 1))
			if err != nil {
				if errors.Is(err, ErrComicsNotFound) {
					return nil
				} else {
					return fmt.Errorf("error get info about comics with id %d: %w", i+1, err)
				}
			}
			ch <- c
			cntGetted.Add(1)
			//ans = append(ans, c)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	start := n
	for int(cntGetted.Load()) < n && start < cntInServer {
		end := start + n - len(ans)
		for ; start < end; start++ {
			g.Go(func() error {
				c, err := xp.GetComicsByID(uint(start + 1))
				if err != nil {
					if errors.Is(err, ErrComicsNotFound) {
						return nil
					} else {
						return fmt.Errorf("error get info about comics with id %d: %w", start+1, err)
					}
				}
				//ans = append(ans, c)
				ch <- c
				cntGetted.Add(1)
				return nil
			})

		}
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	close(ch)
	wg.Wait()
	return ans, nil
}

func (xp XkcdParse) PartParse(isNotExist []uint, cntInServer, n int) ([]Comics, error) {
	ans := make([]Comics, len(isNotExist))
	g := &errgroup.Group{}
	ch := make(chan Comics)
	wg := sync.WaitGroup{}
	wg.Add(1)
	ReturnArrayAfterWriteFromChan(&wg, &ans, ch, 0)
	cnt := atomic.Int32{}
	for _, ID := range isNotExist {
		ID := ID
		g.Go(func() error {
			c, err := xp.GetComicsByID(ID)
			if err != nil {
				if errors.Is(err, ErrComicsNotFound) {
					return nil
				} else {
					return fmt.Errorf("error get info about comics with id %d: %w", ID, err)
				}
			}
			ch <- c
			cnt.Add(1)
			return nil
		})

	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	start := n
	for int(cnt.Load()) < len(isNotExist) {
		end := n + len(isNotExist) - int(cnt.Load())
		for ; start < end && start < cntInServer; start++ {
			start := start
			g.Go(func() error {
				c, err := xp.GetComicsByID(uint(start + 1))
				if err != nil {
					if errors.Is(err, ErrComicsNotFound) {
						return nil
					} else {
						return fmt.Errorf("error get info about comics with id %d: %w", start+1, err)
					}
				}
				ch <- c
				cnt.Add(1)
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	close(ch)
	wg.Wait()
	return ans, nil
}

func ReturnArrayAfterWriteFromChan(wg *sync.WaitGroup, ans *[]Comics, ch <-chan Comics, i int) {
	go func() {
		defer wg.Done()
		for msg := range ch {
			(*ans)[i] = msg
			i++
		}
	}()

}

func (xp XkcdParse) GetComicsByID(ID uint) (Comics, error) {
	resp, err := http.Get(xp.GetURLByID(ID))
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

func (xp XkcdParse) GetURLByID(ID uint) string {
	return fmt.Sprintf("https://%s/%d/info.0.json", xp.URL, ID)
}

func (xp XkcdParse) GetCountComicsInServer() (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://%s/info.0.json", xp.URL))
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
