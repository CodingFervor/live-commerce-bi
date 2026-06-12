package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var secretKey []byte

func SetSecret(secret string) {
	secretKey = []byte(secret)
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwtv5.RegisteredClaims
}

func GenerateToken(userID int64, username, role string, expireHours time.Duration) (string, error) {
	// Generate unique token ID for potential revocation
	jti, err := generateTokenID()
	if err != nil {
		return "", err
	}

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ID:        jti,
			Issuer:    "live-commerce-bi",
			Subject:   username,
			Audience:  jwtv5.ClaimStrings{"live-commerce-bi-api"},
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(expireHours * time.Hour)),
			IssuedAt:  jwtv5.NewNumericDate(time.Now()),
		},
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func generateTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwtv5.ParseWithClaims(tokenStr, &Claims{}, func(t *jwtv5.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
