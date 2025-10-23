# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN adduser -D -s /bin/sh appuser

WORKDIR /root/

COPY --from=builder /app/main .

RUN chown appuser:appuser main

USER appuser

EXPOSE 5465

CMD ["./main"]
