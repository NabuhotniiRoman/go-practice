package handlers

import (
	"net/http"

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
