build:
	go build -o xkcd-server ./cmd/xkcd
bench: build
	./xkcd -c="config.yaml"
	 go test -bench=. ./pkg/database ./pkg/index
