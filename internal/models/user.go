package models

import "time"

// User представляє користувача системи
type User struct {
	ID       string    `json:"id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Picture  string    `json:"picture,omitempty"`
	Sub      string    `json:"sub"`       // OIDC subject identifier
	Iss      string    `json:"iss"`       // OIDC issuer
	Aud      []string  `json:"aud"`       // OIDC audience
	Exp      int64     `json:"exp"`       // OIDC expiration time
	Iat      int64     `json:"iat"`       // OIDC issued at time
	AuthTime int64     `json:"auth_time"` // OIDC authentication time
	CreateAt time.Time `json:"created_at"`
	UpdateAt time.Time `json:"updated_at"`
}

// UserProfile представляє профіль користувача
type UserProfile struct {
	ID       string            `json:"id"`
	Email    string            `json:"email"`
	Name     string            `json:"name"`
	Picture  string            `json:"picture,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// LoginRequest представляє запит на вхід через OIDC
type OIDCLoginRequest struct {
	RedirectURI string `json:"redirect_uri,omitempty"`
	State       string `json:"state,omitempty"`
}

// LoginResponse представляє відповідь на OIDC запит входу
type OIDCLoginResponse struct {
	AuthURL   string `json:"auth_url"`
	State     string `json:"state"`
	SessionID string `json:"session_id"`
}

// LoginRequest представляє запит на вхід через email/password
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse представляє відповідь на успішний вхід
type LoginResponse struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	AccessToken string `json:"access_token"`
	Message     string `json:"message"`
}

// RegisterRequest представляє запит на реєстрацію
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterResponse представляє відповідь на реєстрацію
type RegisterResponse struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Message string `json:"message"`
	AuthURL string `json:"auth_url,omitempty"` // для автоматичного входу після реєстрації
}
