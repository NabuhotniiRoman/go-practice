package handlers

import (
	"go-practice/internal/models"
	"go-practice/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthHandler містить handlers для OIDC authentication
type AuthHandler struct {
	authService        services.AuthService
	postLogoutRedirect string
}

// NewAuthHandler створює новий AuthHandler
func NewAuthHandler(authService services.AuthService, postLogoutRedirect string) *AuthHandler {
	return &AuthHandler{
		authService:        authService,
		postLogoutRedirect: postLogoutRedirect,
	}
}

// @Router /auth/default/login [post]
func (h *AuthHandler) DefaultLogin(c *gin.Context) {
	logrus.Info("🔐 Default Login request")

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid login request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Missing or invalid email/password",
		})
		return
	}

	response, err := h.authService.DefaultLogin(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to login user")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_grant",
			"error_description": "Invalid email or password",
		})
		return
	}

	logrus.WithField("user_id", response.UserID).Info("User logged in successfully")
	c.JSON(http.StatusOK, response)
}

// Login ініціює OIDC Authorization Code Flow
// @Summary OIDC Login
// @Description Ініціює OIDC Authorization Code Flow (Google Login)
// @Tags auth
// @Accept json
// @Produce json
// @Param redirect_uri query string false "Redirect URI"
// @Success 200 {object} models.OIDCLoginResponse
// @Failure 500 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	logrus.Info("🔐 OIDC Login request")

	redirectURI := c.Query("redirect_uri")
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/auth/callback"
	}

	response, err := h.authService.Login(redirectURI)
	if err != nil {
		logrus.WithError(err).Error("Failed to initiate OIDC login")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to initiate OIDC login",
		})
		return
	}

	logrus.WithField("session_id", response.SessionID).Info("OIDC login initiated successfully")
	c.JSON(http.StatusOK, response)
}

// Callback обробляє callback від OIDC провайдера (Authorization Code Flow)
// @Summary OIDC Callback
// @Description Обробляє callback від OIDC провайдера (Authorization Code Flow)
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "Authorization Code"
// @Param state query string true "State"
// @Success 200 {object} models.Token
// @Failure 400 {object} map[string]interface{}
// @Router /auth/callback [get]
func (h *AuthHandler) Callback(c *gin.Context) {
	logrus.Info("🔄 OIDC Authorization Code callback")

	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	// Перевірка на помилки від OIDC провайдера
	if errorParam != "" {
		errorDesc := c.Query("error_description")
		logrus.WithFields(logrus.Fields{
			"error":       errorParam,
			"description": errorDesc,
		}).Error("OIDC provider returned error")

		c.JSON(http.StatusBadRequest, gin.H{
			"error":             errorParam,
			"error_description": errorDesc,
		})
		return
	}

	// Валідація обов'язкових параметрів
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Missing authorization code",
		})
		return
	}

	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Missing state parameter",
		})
		return
	}

	// Використовуємо AuthService для обробки callback
	tokens, user, err := h.authService.HandleCallback(code, state)
	if err != nil {
		logrus.WithError(err).Error("Failed to handle OIDC callback")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_grant",
			"error_description": "Failed to process OIDC callback",
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"code":    code[:10] + "...",
		"state":   state,
		"user_id": user.ID,
	}).Info("OIDC callback processed successfully")

	// Редіректимо клієнта у React додаток
	c.Redirect(http.StatusSeeOther, h.postLogoutRedirect+"?token="+tokens.AccessToken)
}

// Logout завершує сесію користувача (OIDC End Session)
// @Summary Logout
// @Description Завершує сесію користувача (OIDC End Session)
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer Access Token"
// @Param id_token_hint query string false "ID Token Hint"
// @Param post_logout_redirect_uri query string false "Post Logout Redirect URI"
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	logrus.Info("🚪 OIDC Logout request")

	authHeader := c.GetHeader("Authorization")
	idTokenHint := c.Query("id_token_hint")
	postLogoutRedirectURI := c.Query("post_logout_redirect_uri")

	var userID string
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token := authHeader[7:]
		// Можна додати метод в AuthService для отримання userID з токена, якщо потрібно
		user, err := h.authService.GetUserInfo(token)
		if err == nil {
			userID = user.ID
		}
	} else if idTokenHint != "" {
		// Якщо потрібно, додати метод для парсингу id_token_hint
	}

	if userID != "" {
		_ = h.authService.Logout(userID)
		logrus.WithField("user_id", userID).Info("User logged out successfully")
	}

	response := gin.H{
		"message": "Logout successful",
	}
	if postLogoutRedirectURI != "" {
		response["redirect_uri"] = postLogoutRedirectURI
	}
	c.JSON(http.StatusOK, response)
}

// Refresh оновлює access token використовуючи refresh token
// @Summary Refresh Token
// @Description Оновлює access token використовуючи refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refreshToken body models.TokenRefreshRequest true "Refresh Token"
// @Success 200 {object} models.Token
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	logrus.Info("🔄 OIDC Token refresh")

	var req models.TokenRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid refresh request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Missing or invalid refresh_token",
		})
		return
	}

	tokens, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		logrus.WithError(err).Error("Failed to refresh token")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_grant",
			"error_description": "Invalid or expired refresh token",
		})
		return
	}

	logrus.Info("Tokens refreshed successfully")
	c.JSON(http.StatusOK, tokens)
}

// UserInfo повертає інформацію про користувача (OIDC UserInfo endpoint)
// @Summary User Info
// @Description Повертає інформацію про користувача (OIDC UserInfo endpoint)
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Access Token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/userinfo [get]
func (h *AuthHandler) UserInfo(c *gin.Context) {
	logrus.Info("👤 OIDC UserInfo request")

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "Missing or invalid Authorization header",
		})
		return
	}

	accessToken := authHeader[7:]
	user, err := h.authService.GetUserInfo(accessToken)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user info")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "User not found or token invalid",
		})
		return
	}

	userInfo := gin.H{
		"sub":            user.ID,
		"email":          user.Email,
		"name":           user.Name,
		"picture":        user.Picture,
		"email_verified": true,
	}

	c.JSON(http.StatusOK, userInfo)
}

// Register реєструє нового користувача
// @Summary Register
// @Description Реєструє нового користувача
// @Tags auth
// @Accept json
// @Produce json
// @Param registerRequest body models.RegisterRequest true "Registration Data"
// @Success 201 {object} models.RegisterResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	logrus.Info("📝 User registration request")

	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid registration request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"email": req.Email,
		"name":  req.Name,
	}).Info("Processing user registration")

	response, err := h.authService.Register(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to register user")
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Registration failed",
			"details": err.Error(),
		})
		return
	}

	logrus.WithField("user_id", response.UserID).Info("User registered successfully")
	c.JSON(http.StatusCreated, response)
}
