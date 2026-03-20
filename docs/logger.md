# Logging Infrastructure Documentation

This document describes the structured logging and observability setup for the Titik Nol Backend.

## Technology Stack
- **Standard Library**: `log/slog` (Go 1.21+ standard for structured logging).
- **Tracing**: Request ID + User ID propagation via custom middleware and context-aware slog handler.

## Overview
The logging infrastructure is designed to provide end-to-end traceability. Every HTTP request is assigned a unique **Request ID**, which is propagated through the application's context and included in all log entries generated during that request. Authenticated requests also carry the **User ID** automatically.

All log lines include `service` and `env` attributes, enabling Datadog's Unified Service Tagging for filtering, alerting, and service catalog integration.

## Components

### 1. Logger Initialization
Located in `internal/infrastructure/logger/logger.go`.
- Configures global `slog` logger based on `LOG_FORMAT` (`json` or `text`) and `LOG_LEVEL`.
- Attaches `service` and `env` base attributes to every log line (Datadog Unified Service Tagging).
- Enables source location (`AddSource`) in non-production environments for easier debugging.
- Wraps the base handler with a `ContextHandler`.

### 2. Context Handler
Located in `internal/infrastructure/logger/context_handler.go`.
- A wrapper for `slog.Handler` that properly implements `WithAttrs` and `WithGroup` for correct attribute chaining.
- Automatically extracts from `context.Context` and adds to every log record:
  - `request_id` — from the Request ID middleware.
  - `usr.id` — from the Auth middleware (Datadog user tracking convention).

### 3. Request ID Middleware
Located in `internal/delivery/http/middleware/request_id.go`.
- Extracts `X-Request-ID` from incoming headers or generates a new UUID.
- Injects the ID into:
  - Gin Context (for handlers).
  - Outgoing Response Header (`X-Request-ID`).
  - Go Context (`c.Request.Context()`) for propagation to usecases and repositories.

### 4. Auth Middleware — User ID Propagation
Located in `internal/delivery/http/middleware/auth_middleware.go`.
- After successful JWT validation, propagates `user_id` into the request context.
- All downstream `slog.*Context()` calls automatically include `usr.id` in the log output.
- Uses the `response` package for error responses (RFC 7807 compliant).

### 5. HTTP Access Logger
Located in `internal/delivery/http/middleware/logger.go`.
- Logs every HTTP request using Datadog standard attribute names:
  - `http.status_code`, `http.method`, `http.url_details.path`, `http.url_details.queryString`
  - `network.client.ip`, `duration`, `http.response_content_length`, `http.useragent`
- Automatically adjusts log level based on response status:
  - `5xx` → ERROR
  - `4xx` → WARN
  - `2xx/3xx` → INFO
- Gin errors include `error.message` and `error.kind` attributes for Datadog Error Tracking.
- Includes `request_id` and `usr.id` automatically via context.

### 6. GORM / Database Logger
Located in `internal/infrastructure/database/slog_logger.go`.
- A custom GORM logger that bridges `gorm.Interface` to `slog`.
- By using `.WithContext(ctx)` in repositories, database queries are logged with the correct `request_id`.
- Slow query detection with configurable threshold (default: 200ms).

## Datadog Integration Readiness

The logging setup is designed for seamless Datadog log ingestion:

| Feature | Status |
|---------|--------|
| JSON structured output | ✅ `LOG_FORMAT=json` |
| Unified Service Tagging (`service`, `env`) | ✅ Auto-attached |
| Request correlation (`request_id`) | ✅ Auto-injected |
| User tracking (`usr.id`) | ✅ Auto-injected for authenticated requests |
| Standard HTTP attributes | ✅ Datadog naming convention |
| Error tracking attributes | ✅ `error.message`, `error.kind` |
| Log level mapping | ✅ slog levels map directly to Datadog |
| Source location | ✅ Enabled in non-production |

When connecting Datadog, configure the agent to collect logs from stdout with `source: go` and `service: titik-nol-backend`.

## Best Practices for Developers

### Use `Context` in Logging
Always pass the `context.Context` to your logging calls. This ensures that the Request ID and User ID are correctly appended to the log entry.

**Incorrect (IDs won't be tied to request):**
```go
slog.Info("Doing something")
```

**Correct (IDs will be included automatically):**
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

### Error Logging
When logging errors, include the error and a descriptive kind for Datadog Error Tracking:
```go
slog.ErrorContext(ctx, "Failed to create transaction",
    "error", err,
    "error.kind", "database_error",
    "account_id", accountID,
)
```

## Configuration
Controlled via `.env`:
```env
LOG_LEVEL=debug   # debug, info, warn, error
LOG_FORMAT=json   # text, json (use json in production)
```
