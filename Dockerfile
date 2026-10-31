# Multi-stage build for ultra-small container footprint (< 20MB)
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o mutefeed .

# Final minimal runtime container
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/mutefeed .

EXPOSE 8899

ENV PORT=8899
ENV HOST=0.0.0.0

CMD ["./mutefeed", "serve", "--port", "8899", "--host", "0.0.0.0"]
