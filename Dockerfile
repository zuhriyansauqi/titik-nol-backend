# ============================================================
# Stage 1: Build
# ============================================================
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/bin/api cmd/api/main.go

# ============================================================
# Stage 2: Production runtime
# ============================================================
FROM alpine:3.21 AS production

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

COPY --from=builder /app/bin/api .
COPY --from=builder /app/migrations ./migrations

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]

CMD ["./api"]

# ============================================================
# Stage 3: Development (with Air hot-reload)
# ============================================================
FROM golang:1.26.1-alpine AS development

WORKDIR /app

RUN apk add --no-cache git \
    && go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

# Source code mounted as volume at runtime — no COPY needed

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
