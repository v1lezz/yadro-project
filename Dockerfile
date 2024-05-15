FROM golang:latest
RUN mkdir /server
ADD . /server/
WORKDIR /server
RUN go build -o xkcd-server ./cmd/xkcd
CMD ["./xkcd-server -c=config.yaml"]