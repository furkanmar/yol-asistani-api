FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /api .
COPY migrations/ ./migrations/

EXPOSE 8080

CMD ["./api"]
