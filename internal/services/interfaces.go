package services

import (
	"time"

	"go-practice/internal/models"

	"github.com/google/uuid"
)

// AuthService інтерфейс для автентифікації
type AuthService interface {
	DefaultLogin(lr *models.LoginRequest) (*models.LoginResponse, error)
	Register(req *models.RegisterRequest) (*models.RegisterResponse, error)
	Login(redirectURI string) (*models.OIDCLoginResponse, error)
	HandleCallback(code, state string) (*models.Token, *models.User, error)
	Logout(userID string) error
	RefreshToken(refreshToken string) (*models.Token, error)
	GetUserInfo(accessToken string) (*models.User, error)
}

// UserService інтерфейс для роботи з користувачами
type UserService interface {
	GetAllUsers() ([]User, error)
	RegisterUser(req models.RegisterRequest) (*models.RegisterResponse, error)
	GetUserByEmail(email string) (*User, error)
	SearchUsers(query string) ([]User, error)
	GetUserByID(id string) (*User, error)
	ValidatePassword(email, password string) (*User, error)
	UpdateUser(userID string, updates map[string]interface{}) error
	AreFriends(userID, friendID uuid.UUID) (bool, error)
	AddFriend(userID, friendID uuid.UUID) error
	DeleteUser(userID string) error
	GetProfile(userID string) (*models.UserProfile, error)
	CreateOrUpdateFromOIDC(sub, email, name, picture string) (*User, error)
}

// User представляє користувача в базі даних
type User struct {
	ID           string    `gorm:"primaryKey;size:255" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Name         string    `gorm:"not null;size:255" json:"name"`
	PasswordHash string    `gorm:"not null;size:255" json:"-"`
	Picture      string    `gorm:"size:500" json:"picture,omitempty"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SessionService інтерфейс для роботи з сесіями
type SessionService interface {
	CreateSession(userID string, token *models.Token) (*models.Session, error)
	GetSession(sessionID string) (*models.Session, error)
	DeleteSession(sessionID string) error
	ValidateSession(sessionID string) (*models.Session, error)
}
