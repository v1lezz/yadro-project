package database

import (
	"fmt"
	"strings"
)

type Comics struct {
	ImgURL   string   `json:"img_url"`
	Keywords []string `json:"keywords"`
}

func (c Comics) String(ID int) string {
	return fmt.Sprintf("ID: %d\nimg_url: %s\nkeywords: \"%s\"", ID, c.ImgURL, strings.Join(c.Keywords, "\", \""))
}

func CheckComics(data map[int]Comics, cntInServer int) []uint {
	ans := make([]uint, 0, cntInServer-len(data))
	for i := 1; i <= cntInServer; i++ {
		if _, ok := data[i]; !ok {
			ans = append(ans, uint(i))
		}
	}
	return ans
}

//func NewComics(mapComics map[string]interface{}) (Comics, error) {
//	c := Comics{}
//	imgURL, ok := mapComics["img_url"]
//	if !ok {
//		return Comics{}, errors.New("field img_url is empty")
//	}
//	vImgURL, ok := imgURL.(string)
//	if !ok {
//		return Comics{}, errors.New("img url is not string")
//	}
//	c.ImgURL = vImgURL
//	keywords, ok := mapComics["keywords"]
//	if !ok {
//		return Comics{}, errors.New("field keywords is empty")
//	}
//	iKeywords, ok := keywords.([]interface{})
//	if !ok {
//		return Comics{}, errors.New("field keywords is not array")
//	}
//	vKeywords := make([]string, 0, len(iKeywords))
//	for _, keyword := range iKeywords {
//		if v, ok := keyword.(string); ok {
//			vKeywords = append(vKeywords, v)
//		} else {
//			return Comics{}, errors.New("field keywords contains not string")
//		}
//	}
//	c.Keywords = vKeywords
//	return c, nil
//}
