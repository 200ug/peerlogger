FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o crawler .

FROM alpine:latest AS base

RUN apk --no-cache add ca-certificates
RUN adduser -D -s /bin/sh crawler

WORKDIR /app

COPY --from=builder /app/crawler .
# note: config.json should be attached as volume
RUN chown -R crawler:crawler /app
USER crawler

ENTRYPOINT ["./crawler"]
