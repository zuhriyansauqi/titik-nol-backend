package jwt_test

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
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

func TestValidateToken_AlgorithmNone(t *testing.T) {
	svc := newTestService()

	// Craft a token with alg:none to simulate an algorithm confusion attack
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodNone, &jwt.CustomClaims{
		UserID: uuid.New(),
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    "test-issuer",
		},
	})
	tokenString, err := token.SignedString(jwtlib.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = svc.ValidateToken(tokenString)
	assert.Error(t, err)
}
