FROM golang:1.24.2 as builder

WORKDIR /app

COPY go.mod go.sum ./

COPY . .
RUN go mod download

RUN go build -o fin-aggregator-service ./cmd/fin-aggregator-service

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y postgresql-client

COPY --from=builder /app/fin-aggregator-service /usr/local/bin/fin-aggregator-service
COPY --from=builder /go/bin/goose /usr/local/bin/goose

COPY ./migrations ./migrations
COPY ./entrypoint.sh ./entrypoint.sh

RUN chmod +x ./entrypoint.sh

CMD ["./entrypoint.sh"]