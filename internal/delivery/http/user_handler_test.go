package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	handler "github.com/mzhryns/titik-nol-backend/internal/delivery/http"
)

func setupUserRouter(mockUC *mocks.MockUserUsecase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/api/v1")
	handler.NewUserHandler(v1, mockUC)
	return r
}

func TestCreateUser_Success(t *testing.T) {
	mockUC := new(mocks.MockUserUsecase)
	r := setupUserRouter(mockUC)

	user := domain.User{Email: "test@example.com", Name: "Test User"}
	body, _ := json.Marshal(user)

	mockUC.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	mockUC := new(mocks.MockUserUsecase)
	r := setupUserRouter(mockUC)

	user := domain.User{Email: "exists@example.com", Name: "Test"}
	body, _ := json.Marshal(user)

	mockUC.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(domain.ErrEmailAlreadyExists)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestGetByID_Success(t *testing.T) {
	mockUC := new(mocks.MockUserUsecase)
	r := setupUserRouter(mockUC)

	id := uuid.New()
	expected := &domain.User{ID: id, Email: "test@example.com", Name: "Test"}

	mockUC.On("GetByID", mock.Anything, id).Return(expected, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users/"+id.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
}

func TestGetByID_InvalidUUID(t *testing.T) {
	mockUC := new(mocks.MockUserUsecase)
	r := setupUserRouter(mockUC)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetByID_NotFound(t *testing.T) {
	mockUC := new(mocks.MockUserUsecase)
	r := setupUserRouter(mockUC)

	id := uuid.New()
	mockUC.On("GetByID", mock.Anything, id).Return(nil, gorm.ErrRecordNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users/"+id.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFetch_Success(t *testing.T) {
	mockUC := new(mocks.MockUserUsecase)
	r := setupUserRouter(mockUC)

	result := &domain.PaginatedResult{
		Items:      []domain.User{{ID: uuid.New(), Email: "a@example.com"}},
		TotalItems: 1,
		Page:       1,
		PerPage:    20,
		TotalPages: 1,
	}

	mockUC.On("Fetch", mock.Anything, domain.PaginationParams{Page: 1, PerPage: 20}).Return(result, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.NotNil(t, res.Meta)
}

func TestFetch_WithCustomPagination(t *testing.T) {
	mockUC := new(mocks.MockUserUsecase)
	r := setupUserRouter(mockUC)

	result := &domain.PaginatedResult{
		Items:      []domain.User{},
		TotalItems: 0,
		Page:       2,
		PerPage:    5,
		TotalPages: 0,
	}

	mockUC.On("Fetch", mock.Anything, domain.PaginationParams{Page: 2, PerPage: 5}).Return(result, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users?page=2&per_page=5", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFetch_Error(t *testing.T) {
	mockUC := new(mocks.MockUserUsecase)
	r := setupUserRouter(mockUC)

	mockUC.On("Fetch", mock.Anything, domain.PaginationParams{Page: 1, PerPage: 20}).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
