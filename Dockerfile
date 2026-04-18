FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" \
    -o /app/goose github.com/pressly/goose/v3/cmd/goose

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server /app/goose /app/
COPY db/migrations /app/db/migrations
EXPOSE 8080
CMD ["./server"]
