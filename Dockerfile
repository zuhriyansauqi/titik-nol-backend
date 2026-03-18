# Stage 1: Build
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api cmd/api/main.go

# Stage 2: Run
FROM alpine:3.21

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache tzdata

# Copy the binary and config
COPY --from=builder /app/bin/api .
COPY --from=builder /app/.env.example .env

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api"]
