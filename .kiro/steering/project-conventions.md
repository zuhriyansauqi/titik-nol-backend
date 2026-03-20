---
inclusion: always
---

# Titik Nol Backend — Project Conventions

## Tech Stack
- Language: Go 1.22+
- Web Framework: Gin (`github.com/gin-gonic/gin`)
- ORM: GORM with PostgreSQL (`gorm.io/gorm`, `gorm.io/driver/postgres`)
- Auth: Google SSO + JWT (`github.com/golang-jwt/jwt/v5`)
- Testing: `testing` + `github.com/stretchr/testify`
- Logging: `log/slog` (structured, context-aware)
- Config: Viper (`github.com/spf13/viper`)
- Containerization: Docker + Docker Compose

## Architecture — Clean Architecture

```
cmd/api/          → Entrypoint
internal/
  domain/         → Entities, interfaces (Repository, Usecase), domain errors
  usecase/        → Business logic (implements domain interfaces)
  repository/     → Data access (implements domain repository interfaces)
  delivery/http/  → HTTP handlers + middleware (Gin)
  infrastructure/ → Config, database, logger
  pkg/            → Shared packages (jwt, google sso, response helpers)
migrations/       → SQL migration files (golang-migrate)
```

Dependencies point inward: `delivery → usecase → domain ← repository`. The `domain` layer has zero external dependencies.

## API Response Standard

All endpoints MUST use the `internal/pkg/response` package. Never construct raw `gin.H{}` responses in handlers (except `/health`).

### Success
```go
response.Success(c, http.StatusOK, "User retrieved", user)
response.SuccessWithMeta(c, http.StatusOK, "Users fetched", users, meta)
```

### Error (RFC 7807)
```go
response.Error(c, http.StatusBadRequest, "Validation failed", "Email is required", fieldErrors)
response.BadRequest(c, "Invalid input", "The 'name' field cannot be empty")
response.NotFound(c, "User not found")
response.InternalServerError(c, "Something went wrong", err.Error())
```

### Response Shape
```json
// Success
{ "success": true, "message": "...", "data": {...}, "meta": {...} }

// Error
{ "success": false, "message": "...", "error": { "title": "...", "status": 400, "detail": "...", "instance": "/path", "errors": [] } }
```

## Common HTTP Status Codes
| Code | Usage |
|------|-------|
| 200 | Successful request |
| 201 | Resource created |
| 400 | Invalid request |
| 401 | Auth required/failed |
| 404 | Not found |
| 409 | Conflict (duplicate) |
| 422 | Validation failed |
| 429 | Rate limited |
| 500 | Server error |

## Domain Errors
Define domain errors in `internal/domain/errors.go` using `errors.New(...)`. Handlers map these to HTTP status codes.

## Logging

- Always use `slog` with context: `slog.InfoContext(ctx, "message", "key", value)`
- Never use `slog.Info(...)` without context in request-scoped code — the request ID won't be attached.
- Use structured key-value attributes, not formatted strings.
- Levels: DEBUG (dev verbose), INFO (flow events), WARN (recoverable), ERROR (critical).
- Config via `.env`: `LOG_LEVEL` and `LOG_FORMAT` (json for prod, text for dev).

## Git Commits — Conventional Commits

Format: `<type>[optional scope]: <description>`

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

Rules:
- Imperative present tense: "add", not "added"
- No capitalized first letter
- No trailing period
- Example: `feat(api): add user authentication endpoint`

## CLI Commands
- `make build` — Build binary
- `make run` — Run locally
- `make test` — Run all tests
- `make test-v` — Verbose tests
- `make test-cover` — Coverage report
- `make lint` — golangci-lint
- `make docker-up` / `make docker-down` — Docker services
- `make migrate-up` / `make migrate-down` — Migrations
- `make migrate-create name=<name>` — New migration
- `make security` — Run vuln-check + sec-scan
