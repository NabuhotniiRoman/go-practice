package handlers

import (
	"net/http"
	"strings"

	"go-practice/internal/middleware"
	"go-practice/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// APIHandler містить handlers для API endpoints
type APIHandler struct {
	userService services.UserService
}

// NewAPIHandler створює новий APIHandler
func NewAPIHandler(userService services.UserService) *APIHandler {
	return &APIHandler{
		userService: userService,
	}
}

// AddFriend додає користувача в друзі
// @Summary Add Friend
// @Description Додає користувача в друзі
// @Tags api
// @Accept json
// @Produce json
// @Param friend_id body string true "ID користувача, якого додаємо в друзі"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends/add [post]
func (h *APIHandler) AddFriend(c *gin.Context) {
	logrus.Info("AddFriend handler called - маршрут працює!")

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		logrus.Error("Failed to get user ID from context in AddFriend")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user ID from context",
		})
		return
	}

	currentUserID, err := h.userService.GetIDByUserID(userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user ID from UserService")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user ID from UserService",
		})
		return
	}

	var req struct {
		FriendID string `json:"friend_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.FriendID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid or missing friend_id",
		})
		return
	}

	trimmedCurrentUserID := strings.TrimSpace(currentUserID)

	if after, ok := strings.CutPrefix(trimmedCurrentUserID, "usr_"); ok {
		trimmedCurrentUserID = after
	}

	rawID := strings.TrimSpace(req.FriendID)

	// Якщо є префікс "usr_", видаляємо його
	if after, ok := strings.CutPrefix(rawID, "usr_"); ok {
		rawID = after
	}

	// Не можна додати себе в друзі
	if rawID == trimmedCurrentUserID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot add yourself as a friend",
		})
		return
	}

	// Перевіряємо чи вже є в друзях
	isFriend, err := h.userService.AreFriends(trimmedCurrentUserID, rawID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check friendship")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check friendship",
		})
		return
	}
	if isFriend {
		c.JSON(http.StatusConflict, gin.H{
			"error": "User is already your friend",
		})
		return
	}

	// Додаємо в друзі
	err = h.userService.AddFriend(trimmedCurrentUserID, rawID)
	if err != nil {
		logrus.WithError(err).Error("Failed to add friend")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add friend",
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":   trimmedCurrentUserID,
		"friend_id": rawID,
	}).Info("Friend added successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":   "Friend added successfully",
		"friend_id": rawID,
	})
}

// GetFriends повертає список друзів поточного користувача
// @Summary Get Friends
// @Description Повертає список друзів поточного користувача
// @Tags api
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends [get]
func (h *APIHandler) GetFriends(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		logrus.Error("Failed to get user ID from context in GetFriends")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user ID from context",
		})
		return
	}

	trimmedUserID := strings.TrimSpace(userID)
	if after, ok := strings.CutPrefix(trimmedUserID, "usr_"); ok {
		trimmedUserID = after
	}

	friends, err := h.userService.GetFriends(trimmedUserID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", trimmedUserID).Error("Failed to get friends")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve friends",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Friends retrieved successfully",
		"data":    friends,
	})
}

// PublicData повертає публічні дані (без автентифікації)
// @Summary Public Data
// @Description Повертає публічні дані (без автентифікації)
// @Tags api
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/public [get]
func (h *APIHandler) PublicData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Public data - no authentication required",
		"data":    []string{"public1", "public2", "public3"},
	})
}

// ProtectedData повертає захищені дані (потребує автентифікацію)
// @Summary Protected Data
// @Description Повертає захищені дані (потребує автентифікацію)
// @Tags api
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/protected [get]
func (h *APIHandler) ProtectedData(c *gin.Context) {
	// Отримуємо поточного користувача з контексту (додається middleware)
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		logrus.Error("Failed to get user from context in ProtectedData")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user context",
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("User accessed protected data")

	c.JSON(http.StatusOK, gin.H{
		"message": "Protected data - authentication required",
		"data":    []string{"sensitive1", "sensitive2", "sensitive3"},
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
		"timestamp": c.GetTime("request_time"),
	})
}

// UserProfile повертає профіль користувача
// @Summary User Profile
// @Description Повертає профіль користувача
// @Tags api
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/profile [get]
func (h *APIHandler) UserProfile(c *gin.Context) {
	// Отримуємо user ID з контексту
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		logrus.Error("Failed to get user ID from context in UserProfile")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user ID from context",
		})
		return
	}

	// Отримуємо повний профіль користувача через UserService
	profile, err := h.userService.GetProfile(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user profile")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve user profile",
			"details": err.Error(),
		})
		return
	}

	logrus.WithField("user_id", userID).Info("User profile retrieved successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "User profile data",
		"profile": profile,
		"status":  "authenticated",
	})
}

// UpdateProfile дозволяє користувачу оновити свій профіль
// @Summary Update Profile
// @Description Оновлює профіль користувача (тільки name, picture)
// @Tags api
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param updateData body object true "Дані для оновлення профілю"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/profile [put]
func (h *APIHandler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user ID from context",
		})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		logrus.WithError(err).Error("Invalid update profile request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Обмежуємо які поля можна оновлювати
	allowedFields := map[string]bool{
		"name":    true,
		"picture": true,
	}

	filteredUpdates := make(map[string]interface{})
	for key, value := range updateData {
		if allowedFields[key] {
			filteredUpdates[key] = value
		}
	}

	if len(filteredUpdates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "No valid fields to update",
			"allowed_fields": []string{"name", "picture"},
		})
		return
	}

	err := h.userService.UpdateUser(userID, filteredUpdates)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to update user profile")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update profile",
			"details": err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"updates": filteredUpdates,
	}).Info("User profile updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"updates": filteredUpdates,
	})
}

// SearchUsers дозволяє шукати користувачів за ім'ям або email
// @Summary Search Users
// @Description Пошук користувачів за ім'ям або email
// @Tags api
// @Produce json
// @Param q query string true "Пошуковий запит (name або email)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/search [post]
func (h *APIHandler) SearchUsers(c *gin.Context) {
	var request struct {
		UserName string `json:"user_name"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	query := request.UserName
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query 'user_name' is required",
		})
		return
	}

	users, err := h.userService.SearchUsers(query)
	if err != nil {
		logrus.WithError(err).WithField("query", query).Error("Failed to search users")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Users search successful",
		"data":    users,
	})
}

// GetUserByID повертає користувача за його ID
// @Summary Get User By ID
// @Description Повертає користувача за його ID
// @Tags api
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/{id} [get]
func (h *APIHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		logrus.WithError(err).WithField("user_id", id).Error("Failed to get user by ID")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User retrieved successfully",
		"user":    user,
	})
}

// @Router /api/v1/users [get]
func (h *APIHandler) Users(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		logrus.WithError(err).Error("Failed to get users")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Users retrieved successfully",
		"data":    users,
	})
}

// UserData повертає розширені дані користувача
// @Summary User Data
// @Description Повертає розширені дані користувача
// @Tags api
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/user-data [get]
func (h *APIHandler) UserData(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user from context",
		})
		return
	}

	// Приклад розширених даних (можна додати статистику, налаштування тощо)
	userData := gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"name":       user.Name,
			"picture":    user.Picture,
			"is_active":  user.IsActive,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
		"metadata": gin.H{
			"login_count": 0, // TODO: implement login tracking
			"last_login":  nil,
			"preferences": gin.H{},
			"permissions": []string{"read", "write"}, // TODO: implement RBAC
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User data retrieved successfully",
		"data":    userData,
	})
}
