FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod go.sum ./

COPY . .
RUN go mod download

RUN go build -o user-service ./cmd/fin-aggregator-service

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y postgresql-client

COPY --from=builder /app/user-service /usr/local/bin/fin-aggregator-service
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY ./migrations ./migrations
COPY ./migrations.sh ./migrations.sh

RUN chmod +x ./migrations.sh

CMD ["./migrations.sh"]