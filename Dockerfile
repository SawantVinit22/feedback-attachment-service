ARG GO_VERSION=1.24

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src/app

RUN apk add --no-cache ca-certificates

COPY app/go.mod app/go.sum ./
RUN go mod download

COPY app/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/feedback-attachment-service ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/feedback-attachment-service /app/feedback-attachment-service

ENV SERVER_ADDR=:8080
ENV HOME=/home/app

EXPOSE 8080

USER app

ENTRYPOINT ["/app/feedback-attachment-service"]
