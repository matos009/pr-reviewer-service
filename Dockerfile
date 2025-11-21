FROM --platform=linux/arm64 golang:1.23 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o pr-service ./cmd/app


FROM --platform=linux/arm64 alpine:3.19

WORKDIR /app
COPY --from=builder /app/pr-service .

RUN chmod +x /app/pr-service

CMD ["./pr-service"]