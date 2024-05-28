start_database:
	docker run -d -p 5432:5432 -e POSTGRES_USER=v1lezz -e POSTGRES_PASSWORD=1234 -e POSTGRES_DB=comics --name postgres_comics postgres:16.2
build:
	go build -o xkcd-server ./cmd/xkcd
run: build start_database
	/server/xkcd-server
bench: build
	./xkcd -c="config.yaml"
	 go test -bench=. ./pkg/database ./pkg/index
