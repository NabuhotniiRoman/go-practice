package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"go-practice/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// JWTService містить логіку для роботи з JWT токенами
type JWTService interface {
	GenerateTokens(user *User) (*models.Token, error)
	ValidateAccessToken(tokenString string) (*jwt.Token, error)
	ValidateIDToken(tokenString string) (*jwt.Token, error)
	ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error)
	GetUserIDFromToken(tokenString string) (string, error)
	ExtractUserIDFromIDToken(idToken string) (string, error)
}

// jwtService реалізація JWTService
type jwtService struct {
	accessSecret  []byte
	idSecret      []byte
	refreshSecret []byte
}

// NewJWTService створює новий JWT сервіс
func NewJWTService(accessSecret, idSecret, refreshSecret string) JWTService {
	return &jwtService{
		accessSecret:  []byte(accessSecret),
		idSecret:      []byte(idSecret),
		refreshSecret: []byte(refreshSecret),
	}
}

// AccessTokenClaims представляє claims для Access Token
type AccessTokenClaims struct {
	UserID string   `json:"sub"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Scope  []string `json:"scope"`
	jwt.RegisteredClaims
}

// IDTokenClaims представляє claims для ID Token (OIDC)
type IDTokenClaims struct {
	UserID        string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture,omitempty"`
	EmailVerified bool   `json:"email_verified"`
	AuthTime      int64  `json:"auth_time"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims представляє claims для Refresh Token
type RefreshTokenClaims struct {
	UserID    string `json:"sub"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// GenerateTokens генерує Access, ID та Refresh токени
func (j *jwtService) GenerateTokens(user *User) (*models.Token, error) {
	now := time.Now()
	accessExpiry := now.Add(time.Hour)            // 1 година
	idExpiry := now.Add(time.Hour)                // 1 година
	refreshExpiry := now.Add(24 * time.Hour * 30) // 30 днів

	// Генерація Access Token
	accessClaims := AccessTokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Scope:  []string{"openid", "profile", "email"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "oidc-api-server",
			Subject:   user.ID,
			Audience:  []string{"oidc-api-client"},
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateJTI(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Генерація ID Token (OIDC)
	idClaims := IDTokenClaims{
		UserID:        user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Picture:       user.Picture,
		EmailVerified: true,
		AuthTime:      now.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "oidc-api-server",
			Subject:   user.ID,
			Audience:  []string{"oidc-api-client"},
			ExpiresAt: jwt.NewNumericDate(idExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateJTI(),
		},
	}

	idToken := jwt.NewWithClaims(jwt.SigningMethodHS256, idClaims)
	idTokenString, err := idToken.SignedString(j.idSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign ID token: %w", err)
	}

	// Генерація Refresh Token
	refreshClaims := RefreshTokenClaims{
		UserID:    user.ID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "oidc-api-server",
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateJTI(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	logrus.WithField("user_id", user.ID).Info("JWT tokens generated successfully")

	return &models.Token{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		IDToken:      idTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 година в секундах
		ExpiresAt:    accessExpiry,
		Scope:        "openid profile email",
	}, nil
}

// ValidateAccessToken валідує Access Token
func (j *jwtService) ValidateAccessToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.accessSecret, nil
	})
}

// ValidateIDToken валідує ID Token
func (j *jwtService) ValidateIDToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &IDTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.idSecret, nil
	})
}

// ValidateRefreshToken валідує Refresh Token
func (j *jwtService) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.refreshSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid refresh token")
}

// GetUserIDFromToken отримує user ID з Access Token
func (j *jwtService) GetUserIDFromToken(tokenString string) (string, error) {
	token, err := j.ValidateAccessToken(tokenString)
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	return "", fmt.Errorf("invalid token claims")
}

// ExtractUserIDFromIDToken витягує user ID з ID токена
func (j *jwtService) ExtractUserIDFromIDToken(idToken string) (string, error) {
	token, err := jwt.ParseWithClaims(idToken, &IDTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.idSecret, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse ID token: %w", err)
	}

	if claims, ok := token.Claims.(*IDTokenClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	return "", fmt.Errorf("invalid ID token claims")
}

// generateJTI генерує унікальний JWT ID
func generateJTI() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
