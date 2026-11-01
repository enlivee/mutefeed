# Stage 1: Build modern React 19 + TypeScript frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build static Go binary with embedded assets
FROM golang:alpine AS builder
WORKDIR /app
ENV GOTOOLCHAIN=auto

# Cache Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and inject production frontend assets
COPY . .
COPY --from=frontend-builder /app/internal/web/static ./internal/web/static
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o mutefeed .

# Stage 3: Minimal runtime container (< 20MB)
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/mutefeed .

EXPOSE 8899
ENV PORT=8899
ENV HOST=0.0.0.0

CMD ["./mutefeed", "serve"]
