build:
	go build -o xkcd ./cmd/xkcd
bench: build
	./xkcd -c="config.yaml"
	 go test -bench=. ./pkg/database ./pkg/index
