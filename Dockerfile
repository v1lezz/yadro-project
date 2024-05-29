FROM golang:1.22-alpine AS BUILD
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o xkcd-server ./cmd/xkcd

FROM alpine:latest
RUN mkdir /server
COPY --from=BUILD /build/xkcd-server /server/xkcd-server
COPY ./config.yaml /server/config.yaml
COPY ./internal/adapters/repository/migrations/20240514095531_create_comics_index.sql /server/20240514095531_create_comics_index.sql
WORKDIR /server
ENTRYPOINT ["/server/xkcd-server"]