# Start from the official Go base image
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/opengateway

FROM alpine:3

ENV CONFIG_PATH=/app/config.yaml

WORKDIR /app

COPY --from=builder /app/main /app/main
COPY --from=builder /app/cmd/opengateway/config.yaml /app/config.yaml

EXPOSE 4000

CMD ["/app/main"]
