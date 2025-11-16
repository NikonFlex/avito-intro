FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/server ./cmd/pr-reviewer

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/server .

EXPOSE 8080

CMD ["./server"]
