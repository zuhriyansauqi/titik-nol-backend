package jwt_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService() jwt.JWTService {
	return jwt.NewJWTService("test-secret-key", "test-issuer", 3600)
}

func TestGenerateAndValidateToken(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()

	token, err := svc.GenerateToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedID, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, parsedID)
}

func TestValidateToken_Invalid(t *testing.T) {
	svc := newTestService()

	_, err := svc.ValidateToken("not-a-valid-token")
	assert.Error(t, err)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc1 := jwt.NewJWTService("secret-1", "issuer", 3600)
	svc2 := jwt.NewJWTService("secret-2", "issuer", 3600)

	token, err := svc1.GenerateToken(uuid.New())
	require.NoError(t, err)

	_, err = svc2.ValidateToken(token)
	assert.Error(t, err)
}
