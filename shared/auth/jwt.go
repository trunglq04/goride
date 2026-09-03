package auth

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT token payload embedded in every access token.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AccessTokenDuration is the lifetime of an access token.
const AccessTokenDuration = 15 * time.Minute

// RefreshTokenDuration is the lifetime of a refresh token.
const RefreshTokenDuration = 7 * 24 * time.Hour

// LoadPrivateKey reads an RSA private key from a PEM file.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	return privateKey, nil
}

// LoadPublicKey reads an RSA public key from a PEM file.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	return publicKey, nil
}

// SignAccessToken creates a signed JWT access token using RS256.
func SignAccessToken(userID, email, phone, role string, privateKey *rsa.PrivateKey) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Phone:  phone,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "goride-auth",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signedToken, nil
}

// ValidateAccessToken verifies a JWT token string and returns the claims.
func ValidateAccessToken(tokenString string, publicKey *rsa.PublicKey) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
