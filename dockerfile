# Build stage
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Final image
FROM alpine:latest

WORKDIR /root/

RUN apk update && apk upgrade && apk add --no-cache ca-certificates

COPY --from=builder /app/server .
COPY gin-server/.env gin-server/.env


# Inject pod IP via env (used in PeerDiscovery)
ENV MY_POD_IP=""

EXPOSE 8080

ENTRYPOINT ["./server"]