FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

ENV GOPROXY=https://goproxy.io,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o auth-service cmd/auth/main.go

FROM alpine:latest

WORKDIR /app

# Install sqlite dependencies
RUN apk add --no-cache sqlite-libs

COPY --from=builder /app/auth-service .

EXPOSE 50051

CMD ["./auth-service"]
