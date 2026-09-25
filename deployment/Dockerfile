FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o library-api ./cmd


FROM migrate/migrate:v4.19.0 AS migrate


FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/library-api .
COPY --from=builder /app/migrations ./migrations
COPY --from=migrate /migrate /usr/local/bin/migrate

EXPOSE 8080

CMD ["./library-api"]