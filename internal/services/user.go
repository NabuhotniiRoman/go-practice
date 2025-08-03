package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
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

func (s *userService) GetAllUsers() ([]User, error) {
	var users []User
	if err := s.db.Where("is_active = ?", true).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return users, nil
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

// SearchUsers
func (s *userService) SearchUsers(query string) ([]User, error) {
	var users []User
	if err := s.db.Where(
		"is_active = ? AND (LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?))",
		true, "%"+query+"%", "%"+query+"%",
	).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	return users, nil
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

// GetIDByUserID отримує ID користувача за його userID
func (s *userService) GetIDByUserID(userID string) (string, error) {
	var user User
	err := s.db.Select("id").Where("id = ? AND is_active = ?", userID, true).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("user not found")
		}
		return "", fmt.Errorf("failed to get user id: %w", err)
	}
	return user.ID, nil
}

// AreFriends перевіряє чи є користувачі друзями
func (s *userService) AreFriends(userID, friendID string) (bool, error) {
	var exists bool
	err := s.db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM friendships
			WHERE user_id = ? AND friend_id = ?
		)
	`, userID, friendID).Scan(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

// AddFriend додає користувача в друзі
func (s *userService) AddFriend(userID, friendID string) error {
	type Friendship struct {
		UserID    string `gorm:"type:text;not null;index"`
		FriendID  string `gorm:"type:text;not null;index"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	friendship := Friendship{
		UserID:    userID,
		FriendID:  friendID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Спочатку перевіряємо чи вже існує такий зв'язок
	var count int64
	err := s.db.Model(&Friendship{}).Where("user_id = ? AND friend_id = ?", friendship.UserID, friendship.FriendID).Count(&count).Error
	if err != nil {
		return err
	}

	// Якщо вже існує - нічого не робимо
	if count > 0 {
		return nil
	}

	// Інакше додаємо новий зв'язок
	err = s.db.Exec(`
		INSERT INTO friendships (user_id, friend_id, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, friendship.UserID, friendship.FriendID, friendship.CreatedAt, friendship.UpdatedAt).Error

	return err
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

// GetFriends повертає список друзів користувача
func (s *userService) GetFriends(userID string) ([]User, error) {
	var friendIDs []string
	err := s.db.Table("friendships").
		Select("friend_id").
		Where("user_id = ?", userID).
		Scan(&friendIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get friend ids: %w", err)
	}
	if len(friendIDs) == 0 {
		return []User{}, nil
	}

	// 🔧 Повернути префікс "usr_" до friendIDs
	for i, id := range friendIDs {
		friendIDs[i] = "usr_" + strings.TrimSpace(id)
	}

	var friends []User
	err = s.db.
		Where("id IN ? AND is_active = ?", friendIDs, true).
		Find(&friends).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}
	return friends, nil
}
