FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server main.go

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/server ./server

EXPOSE 7777
CMD ["./server"]