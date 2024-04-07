package xkcd

type Comics struct {
	ID         uint   `json:"num"`
	ImgURL     string `json:"img"`
	Transcript string `json:"transcript"`
	Alt        string `json:"title"`
}
