# Logging Infrastructure Documentation

This document describes the structured logging and observability setup for the Titik Nol Backend.

## Technology Stack
- **Standard Library**: `log/slog` (Go 1.21+ standard for structured logging).
- **Tracing**: Request ID propagation via custom middleware and context-aware slog handler.

## Overview
The logging infrastructure is designed to provide end-to-end trace-ability. Every HTTP request is assigned a unique **Request ID**, which is propagated through the application's context and included in all log entries generated during that request.

## Components

### 1. Logger Initialization
Located in `internal/infrastructure/logger/logger.go`.
- Configures global `slog` logger based on `LOG_FORMAT` (`json` or `text`) and `LOG_LEVEL`.
- Wraps the base handler with a `ContextHandler`.

### 2. Context Handler
Integrated into `internal/infrastructure/logger/logger.go`.
- A wrapper for `slog.Handler`.
- Automatically extracts `request_id` from the `context.Context` and adds it as a structured attribute to every log record.

### 3. Request ID Middleware
Located in `internal/delivery/http/middleware/request_id.go`.
- Extracts `X-Request-ID` from incoming headers or generates a new UUID.
- Injects the ID into:
  - Gin Context (for handlers).
  - Outgoing Response Header (`X-Request-ID`).
  - Go Context (`c.Request.Context()`) for propagation to usecases and repositories.

### 4. HTTP Access Logger
Located in `internal/delivery/http/middleware/logger.go`.
- Logs every HTTP request with status, method, latency, and standard attributes.
- Includes the `request_id` in the log record.

### 5. GORM / Database Logger
Located in `internal/infrastructure/database/slog_logger.go`.
- A custom GORM logger that bridges `gorm.Interface` to `slog`.
- By using `.WithContext(ctx)` in repositories, database queries are logged with the correct `request_id`.

## Best Practices for Developers

### Use `Context` in Logging
Always pass the `context.Context` to your logging calls. This ensures that the Request ID is correctly appended to the log entry.

**Incorrect (ID won't be tied to request):**
```go
slog.Info("Doing something")
```

**Correct (ID will be included automatically):**
```go
slog.InfoContext(ctx, "Doing something")
```

### Leveled Logging
- `DEBUG`: Verbose information for development.
- `INFO`: General application flow events (e.g., "User logged in", "Report generated").
- `WARN`: Recoverable issues or unexpected states that don't block the request.
- `ERROR`: Critical failures that require attention (e.g., database connection lost).

### Structured Attributes
Instead of formatting strings, pass attributes as key-value pairs:
```go
slog.InfoContext(ctx, "User action", "user_id", user.ID, "action", "delete")
```

## Configuration
Controlled via `.env`:
```env
LOG_LEVEL=debug   # debug, info, warn, error
LOG_FORMAT=json   # text, json (use json in production)
```
