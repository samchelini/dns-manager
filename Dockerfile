# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY dns ./dns
COPY jsend ./jsend
COPY uuid ./uuid

RUN CGO_ENABLED=0 GOOS=linux go build -o /dns-manager

# Deploy the application binary into a lean image
FROM alpine:latest AS runner

WORKDIR /

COPY --from=builder /dns-manager /dns-manager

EXPOSE 8080

ENTRYPOINT ["/dns-manager"]
