package xkcd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

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

func (xp XkcdParse) FullParse(cntInServer int) ([]Comics, error) {
	g := &errgroup.Group{}
	ans := make([]Comics, 0, cntInServer)
	for i := 0; i < cntInServer; i++ {
		i := i
		g.Go(func() error {
			c, err := xp.GetComicsByID(uint(i + 1))
			if err != nil {
				if errors.Is(err, ErrComicsNotFound) {
					return nil
				} else {
					return fmt.Errorf("error get info about comics with id %d: %w", i+1, err)
				}
			}
			ans = append(ans, c)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return ans, nil
}

func (xp XkcdParse) PartParse(isNotExist []uint) ([]Comics, error) {
	ans := make([]Comics, 0, len(isNotExist))
	for _, ID := range isNotExist {
		c, err := xp.GetComicsByID(ID)
		if err != nil {
			if errors.Is(err, ErrComicsNotFound) {
				continue
			} else {
				return nil, fmt.Errorf("error get info about comics with id %d: %w", ID, err)
			}
		}
		ans = append(ans, c)
	}
	return ans, nil
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
