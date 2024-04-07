package database

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Comics struct {
	ImgURL   string   `json:"img_url"`
	Keywords []string `json:"keywords"`
}

func (c Comics) String(ID int) string {
	//b, err := json.Marshal(c)
	//b, err := json.Marshal(map[string]interface{}{
	//	strconv.Itoa(ID): map[string]interface{}{
	//		"img_url":  c.ImgURL,
	//		"keywords": c.Keywords,
	//	},
	//})
	//if err != nil {
	//	return "", err
	//}
	//return string(b), nil
	return fmt.Sprintf("ID: %d\nimg_url: %s\nkeywords: \"%s\"", ID, c.ImgURL, strings.Join(c.Keywords, "\", \""))
}

func CheckComics(data map[string]Comics, n int) []uint {
	ans := make([]uint, 0, n)
	for i := 1; i <= n; i++ {
		if _, ok := data[strconv.Itoa(i)]; !ok {
			ans = append(ans, uint(i))
		}
	}
	return ans
}

func NewComics(mapComics map[string]interface{}) (Comics, error) {
	c := Comics{}
	imgURL, ok := mapComics["img_url"]
	if !ok {
		return Comics{}, errors.New("field img_url is empty")
	}
	vImgURL, ok := imgURL.(string)
	if !ok {
		return Comics{}, errors.New("img url is not string")
	}
	c.ImgURL = vImgURL
	keywords, ok := mapComics["keywords"]
	if !ok {
		return Comics{}, errors.New("field keywords is empty")
	}
	iKeywords, ok := keywords.([]interface{})
	if !ok {
		return Comics{}, errors.New("field keywords is not array")
	}
	vKeywords := make([]string, 0, len(iKeywords))
	for _, keyword := range iKeywords {
		if v, ok := keyword.(string); ok {
			vKeywords = append(vKeywords, v)
		} else {
			return Comics{}, errors.New("field keywords contains not string")
		}
	}
	c.Keywords = vKeywords
	return c, nil
}
