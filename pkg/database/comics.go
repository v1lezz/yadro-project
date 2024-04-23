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
