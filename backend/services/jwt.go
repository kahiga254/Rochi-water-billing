package services

import (
	"fmt"
	"time"

	"waterbilling/backend/models"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secretKey     string
	tokenDuration time.Duration
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTService(secretKey string, tokenDuration time.Duration) *JWTService {
	return &JWTService{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken generates a JWT token for a user
func (js *JWTService) GenerateToken(user *models.User) (string, error) {
	claims := Claims{
		UserID:   user.ID.Hex(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(js.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.Hex(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(js.secretKey))
}

// GenerateRefreshToken generates a refresh token
func (js *JWTService) GenerateRefreshToken(user *models.User) (string, error) {
	claims := Claims{
		UserID:   user.ID.Hex(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(js.tokenDuration * 24 * 7)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.Hex(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(js.secretKey))
}

// ValidateToken validates a JWT token
func (js *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(js.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken refreshes an expired token
func (js *JWTService) RefreshToken(refreshToken string) (string, error) {
	claims, err := js.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Create new token with same claims but new expiration
	newClaims := &Claims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(js.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   claims.Subject,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return token.SignedString([]byte(js.secretKey))
}

// GetTokenDuration returns the token duration
func (js *JWTService) GetTokenDuration() time.Duration {
	return js.tokenDuration
}
