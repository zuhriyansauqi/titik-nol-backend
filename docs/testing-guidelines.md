# Testing Guidelines

This project follows strict testing conventions for a clean, manageable, and highly readable test suite. All tests must adhere to the following rules to ensure consistency across the backend.

## 1. Top-Level Test Functions (No `t.Run` Subtests)

To maximize clarity, isolate test failures in CI logs, and make executing specific test cases from the CLI straightforward, we **do not use `t.Run()` grouped subtests**. Instead, use discrete, top-level functions for every scenario.

**Incorrect (`t.Run` pattern):**
```go
func TestUserHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) { ... })
	t.Run("duplicate email", func(t *testing.T) { ... })
}
```

**Correct (Top-level function pattern):**
```go
func TestUserHandler_CreateUser_Success(t *testing.T) { ... }

func TestUserHandler_CreateUser_DuplicateEmail(t *testing.T) { ... }
```

## 2. Naming Convention

Test functions must follow the format: `Test[Component]_[Scenario]` or `Test[Function]_[Condition]`.
- **Component / Function**: What is being tested (e.g., `UserHandler`, `AuthMiddleware`, `CreateUser`).
- **Scenario / Condition**: The specific runtime state being verified (e.g., `Success`, `MissingHeader`, `NotFound`, `DuplicateEmail`).

Examples: 
- `TestAuthMiddleware_ValidToken`
- `TestRateLimiter_BlocksRequestsOverLimit`
- `TestGetByID_InvalidUUID`

## 3. Isolated Setup & Tear Down

Each test function should instantiate its own dependencies rather than relying on global package-level variables. Use setup helper functions when necessary, but always return clean interfaces or routers.

```go
// Setup helper
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

## 4. HTTP Recorder Asserts

For delivery layer (HTTP Handler & Middleware) testing, leverage `net/http/httptest` to record responses. Use `github.com/stretchr/testify/assert` and `require` to validate outputs.

```go
w := httptest.NewRecorder()
req, _ := http.NewRequest(http.MethodGet, "/endpoint", nil)
r.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
```

## 5. Mocking Interfaces

All interface dependencies (Repositories, Usecases, Services) should be mocked using `github.com/stretchr/testify/mock`. Define the expected behavior explicitly for every test scenario to guarantee boundaries are respected.
