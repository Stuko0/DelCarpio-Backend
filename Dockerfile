# ── Build stage ──
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /serve ./cmd/serve/

# ── Run stage ──
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /serve .
EXPOSE 8080
ENTRYPOINT ["/app/serve"]
