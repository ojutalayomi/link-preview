# Build stage
FROM golang:1.22-alpine AS builder

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

WORKDIR /home/appuser/

COPY --from=builder /app/link-preview-api .

EXPOSE 5465

CMD ["./link-preview-api"]
