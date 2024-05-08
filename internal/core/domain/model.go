package domain

import (
	"fmt"
	"strings"
)

type Comics struct {
	ID       int      `json:"-"`
	ImgURL   string   `json:"img_url"`
	Keywords []string `json:"keywords,omitempty"`
}

type UpdateMeta struct {
	New   int `json:"new"`
	Total int `json:"total"`
}

func (c Comics) String() string {
	return fmt.Sprintf("ID: %d\nimg_url: %s\nkeywords: \"%s\"", c.ID, c.ImgURL, strings.Join(c.Keywords, "\", \""))
}
