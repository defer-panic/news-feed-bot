FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY internal ./internal
COPY cmd ./cmd

RUN go build -o /app/news-feed-bot ./cmd/

EXPOSE 8080

CMD ["/app/news-feed-bot"]