# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api cmd/api/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/bin/api .
COPY --from=builder /app/migrations ./migrations

EXPOSE 10000

CMD ["./api"]
