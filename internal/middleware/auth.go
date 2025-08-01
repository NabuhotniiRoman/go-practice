package middleware

import (
	"net/http"
	"strings"

	"go-practice/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware створює middleware для перевірки JWT токенів
func AuthMiddleware(jwtService services.JWTService, userService services.UserService) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Отримуємо Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logrus.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "unauthorized",
				"error_description": "Missing Authorization header",
			})
			c.Abort()
			return
		}

		// Перевіряємо формат Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logrus.Warn("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "unauthorized",
				"error_description": "Invalid Authorization header format",
			})
			c.Abort()
			return
		}

		// Витягуємо токен
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			logrus.Warn("Empty access token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "unauthorized",
				"error_description": "Empty access token",
			})
			c.Abort()
			return
		}

		// Валідуємо токен через JWTService
		userID, err := jwtService.GetUserIDFromToken(token)
		if err != nil {
			logrus.WithError(err).Warn("Invalid access token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "invalid_token",
				"error_description": "Token validation failed",
			})
			c.Abort()
			return
		}

		// Отримуємо користувача з бази даних
		user, err := userService.GetUserByID(userID)
		if err != nil {
			logrus.WithError(err).Error("Failed to get user from token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "invalid_token",
				"error_description": "User not found",
			})
			c.Abort()
			return
		}

		// Перевіряємо чи користувач активний
		if !user.IsActive {
			logrus.WithField("user_id", userID).Warn("Inactive user attempted access")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "account_disabled",
				"error_description": "User account is disabled",
			})
			c.Abort()
			return
		}

		// Зберігаємо користувача в контексті для подальшого використання
		c.Set("user", user)
		c.Set("user_id", userID)

		logrus.WithFields(logrus.Fields{
			"user_id": userID,
			"email":   user.Email,
			"path":    c.Request.URL.Path,
		}).Debug("User authenticated successfully")

		// Продовжуємо обробку запиту
		c.Next()
	})
}

// GetCurrentUser витягує поточного користувача з контексту
func GetCurrentUser(c *gin.Context) (*services.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userObj, ok := user.(*services.User)
	return userObj, ok
}

// GetCurrentUserID витягує ID поточного користувача з контексту
func GetCurrentUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	return userIDStr, ok
}
