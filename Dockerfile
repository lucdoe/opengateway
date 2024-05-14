# Start from the official Go base image
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/gateway/main.go

FROM alpine:latest  

ENV GO_ENV=production
ENV CONFIG_PATH=/app/config.yaml

WORKDIR /app

COPY --from=builder /app/main /app/main
COPY --from=builder /app/cmd/gateway/config.yaml /app/config.yaml

RUN chmod +x /app/main

EXPOSE 4000

CMD ["/app/main"]
