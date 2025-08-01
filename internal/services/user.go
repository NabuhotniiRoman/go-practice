package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"go-practice/internal/models"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// userService реалізація UserService
type userService struct {
	db *gorm.DB
}

// NewUserService створює новий UserService
func NewUserService(db *gorm.DB) UserService {
	return &userService{
		db: db,
	}
}

// RegisterUser реєструє нового користувача
func (s *userService) RegisterUser(req models.RegisterRequest) (*models.RegisterResponse, error) {
	// Перевіряємо чи користувач вже існує
	var existingUser User
	err := s.db.Where("email = ?", req.Email).First(&existingUser).Error
	if err == nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Хешуємо пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Генеруємо унікальний ID
	userID, err := generateUserID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}

	// Створюємо нового користувача
	user := User{
		ID:           userID,
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Зберігаємо в базу даних
	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Повертаємо відповідь
	response := &models.RegisterResponse{
		UserID:  user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Message: "User registered successfully",
	}

	return response, nil
}

// GetUserByEmail отримує користувача за email
func (s *userService) GetUserByEmail(email string) (*User, error) {
	var user User
	err := s.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByID отримує користувача за ID
func (s *userService) GetUserByID(id string) (*User, error) {
	var user User
	err := s.db.Where("id = ? AND is_active = ?", id, true).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// ValidatePassword перевіряє пароль користувача
func (s *userService) ValidatePassword(email, password string) (*User, error) {
	user, err := s.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

// UpdateUser оновлює дані користувача
func (s *userService) UpdateUser(userID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	result := s.db.Model(&User{}).Where("id = ? AND is_active = ?", userID, true).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// DeleteUser деактивує користувача (soft delete)
func (s *userService) DeleteUser(userID string) error {
	result := s.db.Model(&User{}).Where("id = ?", userID).Update("is_active", false)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// GetProfile повертає профіль користувача
func (s *userService) GetProfile(userID string) (*models.UserProfile, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return &models.UserProfile{
		ID:      user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Picture: user.Picture,
	}, nil
}

// CreateOrUpdateFromOIDC створює нового користувача або оновлює існуючого на основі даних від OIDC провайдера
func (s *userService) CreateOrUpdateFromOIDC(sub, email, name, picture string) (*User, error) {
	logrus.WithFields(logrus.Fields{
		"sub":   sub,
		"email": email,
		"name":  name,
	}).Info("Creating or updating user from OIDC provider")

	// Спробуємо знайти користувача за email
	existingUser, err := s.GetUserByEmail(email)
	if err == nil {
		// Користувач існує, оновлюємо дані
		updates := map[string]interface{}{
			"name":    name,
			"picture": picture,
		}

		if err := s.UpdateUser(existingUser.ID, updates); err != nil {
			return nil, fmt.Errorf("failed to update existing user: %w", err)
		}

		// Повертаємо оновленого користувача
		return s.GetUserByID(existingUser.ID)
	}

	// Користувач не існує, створюємо нового
	userID, err := generateUserID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}

	newUser := User{
		ID:           userID,
		Email:        email,
		Name:         name,
		Picture:      picture,
		PasswordHash: "", // Для OIDC користувачів пароль не потрібен
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.Create(&newUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create user from OIDC: %w", err)
	}

	logrus.WithField("user_id", newUser.ID).Info("User created successfully from OIDC provider")
	return &newUser, nil
}

// generateUserID генерує унікальний ID для користувача
func generateUserID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "usr_" + hex.EncodeToString(bytes), nil
}
