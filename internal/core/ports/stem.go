package ports

type Stemmer interface {
	Stem([]string) ([]string, error)
}
