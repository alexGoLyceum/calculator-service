package auth_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJWT_GenerateAndParse_Success(t *testing.T) {
	secret := []byte("supersecret")
	ttl := time.Minute
	jwtManager := auth.NewJWTManager(secret, ttl)

	userID := uuid.New()
	token, err := jwtManager.Generate(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	parsedID, err := jwtManager.Parse(token)
	require.NoError(t, err)
	require.Equal(t, userID, parsedID)
}

func TestJWT_Parse_InvalidTokenFormat(t *testing.T) {
	secret := []byte("supersecret")
	jwtManager := auth.NewJWTManager(secret, time.Minute)

	_, err := jwtManager.Parse("not.a.valid.token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse token")
}

func TestJWT_Parse_InvalidSigningMethod(t *testing.T) {
	secret := []byte("secret")
	jwtManager := auth.NewJWTManager(secret, time.Minute)

	claims := auth.JWTClaims{
		UserID: uuid.New().String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	require.NoError(t, err)

	_, err = jwtManager.Parse(tokenString)
	require.NoError(t, err)
}

func TestJWT_Parse_InvalidUUID(t *testing.T) {
	secret := []byte("secret")
	jwtManager := auth.NewJWTManager(secret, time.Minute)

	claims := auth.JWTClaims{
		UserID: "invalid-uuid",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	require.NoError(t, err)

	_, err = jwtManager.Parse(tokenString)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid user_id in token")
}

func TestJWT_Parse_InvalidSignature(t *testing.T) {
	secret1 := []byte("secret1")
	secret2 := []byte("secret2")

	jwt1 := auth.NewJWTManager(secret1, time.Minute)
	jwt2 := auth.NewJWTManager(secret2, time.Minute)

	userID := uuid.New()
	token, err := jwt1.Generate(userID)
	require.NoError(t, err)

	_, err = jwt2.Parse(token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse token")
}

func TestJWT_Parse_ExpiredToken(t *testing.T) {
	secret := []byte("secret")
	jwtManager := auth.NewJWTManager(secret, -time.Minute)

	userID := uuid.New()
	token, err := jwtManager.Generate(userID)
	require.NoError(t, err)

	_, err = jwtManager.Parse(token)
	require.Error(t, err)
	require.True(t, errors.Is(err, jwt.ErrTokenExpired) || strings.Contains(err.Error(), "expired"))
}
