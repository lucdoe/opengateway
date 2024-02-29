# syntax=docker/dockerfile:1

FROM golang:alpine3.19 AS builder

# Install git and bash. Bash is required for wait-for-it.sh
RUN apk update && apk add --no-cache git bash

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY cmd/*.go ./cmd/
COPY app/*.go ./app/
COPY internal ./internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-go ./cmd

# Final stage
FROM alpine:3.19

# Install bash for wait-for-it.sh
RUN apk add --no-cache bash

WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /docker-go .

# Add wait-for-it
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

# Optional: Expose port
EXPOSE 8080

# Use wait-for-it.sh to wait for the postgres service to be available before starting the application
CMD ["/wait-for-it.sh", "postgres:5432", "--", "./docker-go"]
