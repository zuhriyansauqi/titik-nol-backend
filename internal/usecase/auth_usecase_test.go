package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	domainmocks "github.com/mzhryns/titik-nol-backend/internal/domain/mocks"
	googlemocks "github.com/mzhryns/titik-nol-backend/internal/pkg/google/mocks"
	jwtmocks "github.com/mzhryns/titik-nol-backend/internal/pkg/jwt/mocks"
	"github.com/mzhryns/titik-nol-backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/idtoken"
)

func newGooglePayload(email, name, picture, subject string) *idtoken.Payload {
	return &idtoken.Payload{
		Subject: subject,
		Claims: map[string]any{
			"email":   email,
			"name":    name,
			"picture": picture,
		},
	}
}

func TestLoginWithGoogle_NewUser(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)
	mockGoogle := new(googlemocks.MockGoogleSSOService)
	mockJWT := new(jwtmocks.MockJWTService)
	uc := usecase.NewAuthUsecase(mockRepo, mockGoogle, mockJWT)

	payload := newGooglePayload("new@example.com", "New User", "http://pic.url", "google-123")
	req := &domain.GoogleLoginRequest{IDToken: "valid-token"}

	mockGoogle.On("VerifyIDToken", mock.Anything, "valid-token").Return(payload, nil)
	mockRepo.On("GetByProviderID", mock.Anything, "google-123").Return(nil, errors.New("not found"))
	mockRepo.On("GetByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("not found"))
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	mockJWT.On("GenerateToken", mock.AnythingOfType("uuid.UUID")).Return("jwt-token", nil)

	resp, err := uc.LoginWithGoogle(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "jwt-token", resp.AccessToken)
	assert.True(t, resp.IsNewUser)
	mockRepo.AssertExpectations(t)
}

func TestLoginWithGoogle_ExistingByProvider(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)
	mockGoogle := new(googlemocks.MockGoogleSSOService)
	mockJWT := new(jwtmocks.MockJWTService)
	uc := usecase.NewAuthUsecase(mockRepo, mockGoogle, mockJWT)

	existingUser := &domain.User{ID: uuid.New(), Email: "existing@example.com", ProviderID: "google-123"}
	payload := newGooglePayload("existing@example.com", "Existing", "http://pic.url", "google-123")
	req := &domain.GoogleLoginRequest{IDToken: "valid-token"}

	mockGoogle.On("VerifyIDToken", mock.Anything, "valid-token").Return(payload, nil)
	mockRepo.On("GetByProviderID", mock.Anything, "google-123").Return(existingUser, nil)
	mockJWT.On("GenerateToken", existingUser.ID).Return("jwt-token", nil)

	resp, err := uc.LoginWithGoogle(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "jwt-token", resp.AccessToken)
	assert.False(t, resp.IsNewUser)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestLoginWithGoogle_ExistingByEmail(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)
	mockGoogle := new(googlemocks.MockGoogleSSOService)
	mockJWT := new(jwtmocks.MockJWTService)
	uc := usecase.NewAuthUsecase(mockRepo, mockGoogle, mockJWT)

	existingUser := &domain.User{ID: uuid.New(), Email: "user@example.com"}
	payload := newGooglePayload("user@example.com", "User", "http://pic.url", "google-456")
	req := &domain.GoogleLoginRequest{IDToken: "valid-token"}

	mockGoogle.On("VerifyIDToken", mock.Anything, "valid-token").Return(payload, nil)
	mockRepo.On("GetByProviderID", mock.Anything, "google-456").Return(nil, errors.New("not found"))
	mockRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(existingUser, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	mockJWT.On("GenerateToken", existingUser.ID).Return("jwt-token", nil)

	resp, err := uc.LoginWithGoogle(context.Background(), req)

	require.NoError(t, err)
	assert.False(t, resp.IsNewUser)
	mockRepo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("*domain.User"))
}

func TestLoginWithGoogle_MissingEmailClaim(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)
	mockGoogle := new(googlemocks.MockGoogleSSOService)
	mockJWT := new(jwtmocks.MockJWTService)
	uc := usecase.NewAuthUsecase(mockRepo, mockGoogle, mockJWT)

	payload := &idtoken.Payload{
		Subject: "google-789",
		Claims:  map[string]any{"name": "User"},
	}
	req := &domain.GoogleLoginRequest{IDToken: "valid-token"}

	mockGoogle.On("VerifyIDToken", mock.Anything, "valid-token").Return(payload, nil)

	resp, err := uc.LoginWithGoogle(context.Background(), req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email")
}

func TestLoginWithGoogle_VerifyFails(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)
	mockGoogle := new(googlemocks.MockGoogleSSOService)
	mockJWT := new(jwtmocks.MockJWTService)
	uc := usecase.NewAuthUsecase(mockRepo, mockGoogle, mockJWT)

	req := &domain.GoogleLoginRequest{IDToken: "bad-token"}
	mockGoogle.On("VerifyIDToken", mock.Anything, "bad-token").Return(nil, errors.New("invalid token"))

	resp, err := uc.LoginWithGoogle(context.Background(), req)

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestGetCurrentUser(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)
	mockGoogle := new(googlemocks.MockGoogleSSOService)
	mockJWT := new(jwtmocks.MockJWTService)
	uc := usecase.NewAuthUsecase(mockRepo, mockGoogle, mockJWT)

	id := uuid.New()
	expected := &domain.User{ID: id, Email: "test@example.com"}
	mockRepo.On("GetByID", mock.Anything, id).Return(expected, nil)

	user, err := uc.GetCurrentUser(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, expected.Email, user.Email)
}
