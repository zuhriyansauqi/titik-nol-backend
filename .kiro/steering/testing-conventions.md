---
inclusion: fileMatch
fileMatchPattern: "**/*_test.go"
---

# Titik Nol Backend — Testing Conventions

## 1. Top-Level Test Functions Only

Do NOT use `t.Run()` subtests. Every scenario gets its own top-level function.

```go
// ❌ Wrong
func TestUserHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) { ... })
	t.Run("duplicate email", func(t *testing.T) { ... })
}

// ✅ Correct
func TestUserHandler_CreateUser_Success(t *testing.T) { ... }
func TestUserHandler_CreateUser_DuplicateEmail(t *testing.T) { ... }
```

## 2. Naming Convention

Format: `Test[Component]_[Scenario]` or `Test[Function]_[Condition]`

Examples:
- `TestAuthMiddleware_ValidToken`
- `TestRateLimiter_BlocksRequestsOverLimit`
- `TestGetByID_InvalidUUID`

## 3. Isolated Setup

Each test instantiates its own dependencies. No shared global state. Use setup helpers that return clean instances.

```go
func setupTestRouter(mockUC *mocks.MockUsecase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler.NewHandler(r, mockUC)
	return r
}

func TestExample_Success(t *testing.T) {
	mockUC := new(mocks.MockUsecase)
	r := setupTestRouter(mockUC)
	// ... test logic
}
```

## 4. HTTP Handler Tests

Use `net/http/httptest` for recording responses. Assert with `testify/assert` and `testify/require`.

```go
w := httptest.NewRecorder()
req, _ := http.NewRequest(http.MethodGet, "/endpoint", nil)
r.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
```

## 5. Mocking

Mock all interface dependencies using `github.com/stretchr/testify/mock`. Mocks live in `mocks/` subdirectories next to the interface definition (e.g., `internal/domain/mocks/`).

Set explicit expectations for every test scenario:
```go
mockUC.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)
```

## 6. Running Tests

- `make test` — all tests
- `make test-v` — verbose
- `make test-cover` — with coverage report
