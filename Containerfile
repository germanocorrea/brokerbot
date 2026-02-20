FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o brokerbot -ldflags="-s -w" .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN adduser -D -H -h /app appuser

WORKDIR /app

COPY --from=builder /app/brokerbot .

RUN chown -R appuser:appuser /app

RUN mkdir -p /run/brokerbot

USER appuser

ENTRYPOINT ["/app/brokerbot"]
