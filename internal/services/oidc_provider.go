package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-practice/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// OIDCProviderService інтерфейс для роботи з зовнішнім OIDC провайдером
type OIDCProviderService interface {
	ExchangeCodeForTokens(code, redirectURI string) (*models.Token, error)
	ValidateIDToken(idToken string) (*IDTokenClaims, error)
	GetUserInfoFromProvider(accessToken string) (*ProviderUserInfo, error)
}

// ProviderUserInfo представляє інформацію про користувача від OIDC провайдера
type ProviderUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture,omitempty"`
	EmailVerified bool   `json:"email_verified"`
}

// TokenResponse представляє відповідь від OIDC провайдера на обмін коду
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
}

// oidcProviderService реалізація OIDCProviderService
type oidcProviderService struct {
	clientID     string
	clientSecret string
	tokenURL     string
	userInfoURL  string
	issuer       string
	httpClient   *http.Client
}

// NewOIDCProviderService створює новий OIDC Provider сервіс
func NewOIDCProviderService(clientID, clientSecret, tokenURL, userInfoURL, issuer string) OIDCProviderService {
	return &oidcProviderService{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     tokenURL,
		userInfoURL:  userInfoURL,
		issuer:       issuer,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExchangeCodeForTokens обмінює authorization code на токени з OIDC провайдером
func (o *oidcProviderService) ExchangeCodeForTokens(code, redirectURI string) (*models.Token, error) {
	logrus.WithFields(logrus.Fields{
		"code":         code[:10] + "...",
		"redirect_uri": redirectURI,
		"token_url":    o.tokenURL,
	}).Info("Exchanging authorization code for tokens")

	// Підготовка параметрів для POST запиту
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", o.clientID)
	data.Set("client_secret", o.clientSecret)

	// Створення HTTP запиту
	req, err := http.NewRequest("POST", o.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Відправка запиту
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for tokens: %w", err)
	}
	defer resp.Body.Close()

	// Читання відповіді
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("OIDC provider returned error")
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Парсинг JSON відповіді
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"token_type": tokenResp.TokenType,
		"expires_in": tokenResp.ExpiresIn,
		"scope":      tokenResp.Scope,
	}).Info("Successfully received tokens from OIDC provider")

	// Конвертація в наш формат
	token := &models.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scope:        tokenResp.Scope,
	}

	return token, nil
}

// ValidateIDToken валідує ID Token від OIDC провайдера
func (o *oidcProviderService) ValidateIDToken(idToken string) (*IDTokenClaims, error) {
	logrus.WithField("id_token", idToken[:20]+"...").Info("Validating ID token from OIDC provider")

	// В реальному застосунку тут має бути:
	// 1. Отримання публічних ключів провайдера з /.well-known/jwks_uri
	// 2. Валідація підпису JWT
	// 3. Валідація issuer, audience, expiration тощо

	// Для демонстрації парсимо токен без валідації підпису
	token, err := jwt.ParseWithClaims(idToken, &IDTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// В продакшені тут має бути правильний ключ від провайдера
		return []byte("dummy-key-for-demo"), nil
	})

	if err != nil {
		// Якщо не вдається розпарсити, спробуємо витягти claims без валідації
		logrus.WithError(err).Warn("Failed to validate ID token signature, attempting to parse claims only")

		// Розділяємо JWT на частини
		parts := strings.Split(idToken, ".")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid ID token format")
		}

		// Декодуємо payload (друга частина)
		payload := parts[1]
		// Додаємо padding якщо потрібно
		for len(payload)%4 != 0 {
			payload += "="
		}

		claims := &IDTokenClaims{}
		_, _, err := jwt.NewParser().ParseUnverified(idToken, claims)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ID token claims: %w", err)
		}

		// Базова валідація
		if claims.Issuer != o.issuer {
			logrus.WithFields(logrus.Fields{
				"expected_issuer": o.issuer,
				"actual_issuer":   claims.Issuer,
			}).Warn("ID token issuer mismatch")
		}

		if time.Now().After(claims.ExpiresAt.Time) {
			return nil, fmt.Errorf("ID token has expired")
		}

		logrus.WithFields(logrus.Fields{
			"sub":   claims.UserID,
			"email": claims.Email,
			"name":  claims.Name,
		}).Info("ID token parsed successfully")

		return claims, nil
	}

	if claims, ok := token.Claims.(*IDTokenClaims); ok && token.Valid {
		logrus.WithFields(logrus.Fields{
			"sub":   claims.UserID,
			"email": claims.Email,
			"name":  claims.Name,
		}).Info("ID token validated successfully")
		return claims, nil
	}

	return nil, fmt.Errorf("invalid ID token claims")
}

// GetUserInfoFromProvider отримує інформацію про користувача з UserInfo endpoint
func (o *oidcProviderService) GetUserInfoFromProvider(accessToken string) (*ProviderUserInfo, error) {
	logrus.WithField("access_token", accessToken[:20]+"...").Info("Getting user info from OIDC provider")

	// Створення HTTP запиту до UserInfo endpoint
	req, err := http.NewRequest("GET", o.userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	// Відправка запиту
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status %d", resp.StatusCode)
	}

	// Читання і парсинг відповіді
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read userinfo response: %w", err)
	}

	var userInfo ProviderUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse userinfo response: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"sub":   userInfo.Sub,
		"email": userInfo.Email,
		"name":  userInfo.Name,
	}).Info("Successfully retrieved user info from OIDC provider")

	return &userInfo, nil
}
