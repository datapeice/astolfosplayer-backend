FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

ENV GOPROXY=https://goproxy.io,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o file-service cmd/file/main.go

FROM alpine:latest

WORKDIR /app

# Install sqlite dependencies
RUN apk add --no-cache sqlite-libs

COPY --from=builder /app/file-service .

EXPOSE 50052

CMD ["./file-service"]
