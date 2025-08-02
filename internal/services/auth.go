package services

import (
	"go-practice/internal/models"

	"github.com/sirupsen/logrus"
)

// authService реалізація AuthService
type authService struct {
	userService         UserService
	jwtService          JWTService
	stateService        StateService
	oidcProviderService OIDCProviderService
	sessionManager      SessionManager
}

// NewAuthService створює новий AuthService
func NewAuthService(userService UserService, jwtService JWTService, stateService StateService, oidcProviderService OIDCProviderService, sessionManager SessionManager) AuthService {
	return &authService{
		userService:         userService,
		jwtService:          jwtService,
		stateService:        stateService,
		oidcProviderService: oidcProviderService,
		sessionManager:      sessionManager,
	}
}

// Register реєструє нового користувача
func (s *authService) Register(req *models.RegisterRequest) (*models.RegisterResponse, error) {
	logrus.WithFields(logrus.Fields{
		"email": req.Email,
		"name":  req.Name,
	}).Info("AuthService: Register called")

	// Використовуємо UserService для реєстрації
	response, err := s.userService.RegisterUser(*req)
	if err != nil {
		logrus.WithError(err).Error("Failed to register user")
		return nil, err
	}

	logrus.WithField("user_id", response.UserID).Info("User registered successfully via AuthService")
	return response, nil
}

func (s *authService) DefaultLogin(lr *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userService.ValidatePassword(lr.Email, lr.Password)
	if err != nil {
		logrus.WithError(err).Error("Failed to validate password")
		return nil, err
	}

	// Генеруємо токени для користувача
	tokens, err := s.jwtService.GenerateTokens(user)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate tokens")
		return nil, err
	}

	// Створюємо сесію для користувача
	_, err = s.sessionManager.CreateSession(user.ID, tokens.AccessToken, tokens.RefreshToken)
	if err != nil {
		logrus.WithError(err).Error("Failed to create session")
		return nil, err
	}

	response := &models.LoginResponse{
		UserID:      user.ID,
		Email:       user.Email,
		Name:        user.Name,
		AccessToken: tokens.AccessToken,
		Message:     "Login successful",
	}

	logrus.Info("User logged in successfully")
	return response, nil
}

func (s *authService) Login(redirectURI string) (*models.OIDCLoginResponse, error) {
	logrus.Info("AuthService: Login called")

	// Створюємо сесію для відстеження OIDC flow
	session, err := s.sessionManager.CreateSession("", "", "") // UserID буде оновлений після успішної автентифікації
	if err != nil {
		logrus.WithError(err).Error("Failed to create session")
		return nil, err
	}

	// Генеруємо state для CSRF захисту, використовуючи session ID
	state, err := s.stateService.GenerateState(session.SessionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate state")
		return nil, err
	}

	// Формуємо URL для OIDC провайдера (приклад для Google)
	redirectURI = "https://api.example.com/auth/callback"

	authURL := "https://accounts.google.com/o/oauth2/v2/auth" +
		"?client_id=906808629445-iakp5ilfkc9ltmnk5j3o001dvvres0tn.apps.googleusercontent.com" +
		"&redirect_uri=" + redirectURI +
		"&scope=openid+profile+email" +
		"&response_type=code" +
		"&state=" + state

	logrus.WithFields(logrus.Fields{
		"state":        state[:10] + "...",
		"session_id":   session.SessionID,
		"redirect_uri": redirectURI,
	}).Info("Generated OIDC login URL with session tracking")

	return &models.OIDCLoginResponse{
		AuthURL:   authURL,
		State:     state,
		SessionID: session.SessionID,
	}, nil
}

// HandleCallback обробляє callback від OIDC провайдера
func (s *authService) HandleCallback(code, state string) (*models.Token, *models.User, error) {
	logrus.WithFields(logrus.Fields{
		"code":  code[:10] + "...",
		"state": state[:10] + "...",
	}).Info("AuthService: HandleCallback called")

	// Валідуємо state для CSRF захисту та отримуємо session ID
	sessionID, err := s.stateService.ValidateState(state)
	if err != nil {
		logrus.WithError(err).Error("State validation failed")
		return nil, nil, err
	}

	// Перевіряємо чи існує сесія
	session, err := s.sessionManager.GetSession(sessionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get session")
		return nil, nil, err
	}
	if session == nil {
		logrus.Error("Session not found or expired")
		return nil, nil, err
	}

	// Обмінюємо authorization code на токени з OIDC провайдера
	providerTokens, err := s.oidcProviderService.ExchangeCodeForTokens(code, "https://api.example.com/auth/callback")
	if err != nil {
		logrus.WithError(err).Error("Failed to exchange code for tokens")
		return nil, nil, err
	}

	// Валідуємо ID token від провайдера
	idTokenClaims, err := s.oidcProviderService.ValidateIDToken(providerTokens.IDToken)
	if err != nil {
		logrus.WithError(err).Error("ID token validation failed")
		return nil, nil, err
	}

	// Створюємо або оновлюємо користувача в нашій системі
	user, err := s.userService.CreateOrUpdateFromOIDC(
		idTokenClaims.UserID,
		idTokenClaims.Email,
		idTokenClaims.Name,
		idTokenClaims.Picture,
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to create/update user from OIDC")
		return nil, nil, err
	}

	// Оновлюємо сесію з user ID
	err = s.sessionManager.UpdateSessionUser(sessionID, user.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to update session with user ID")
	}

	// Генеруємо наші внутрішні JWT токени
	tokens, err := s.jwtService.GenerateTokens(user)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate internal tokens")
		return nil, nil, err
	}

	// Конвертуємо user в models.User
	modelUser := &models.User{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Picture:  user.Picture,
		CreateAt: user.CreatedAt,
		UpdateAt: user.UpdatedAt,
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    user.ID,
		"session_id": sessionID,
	}).Info("OIDC callback processed successfully with session tracking")

	return tokens, modelUser, nil
}

// Logout завершує сесію користувача
func (s *authService) Logout(userID string) error {
	logrus.WithField("userID", userID).Info("AuthService: Logout called")

	// Перевіряємо чи користувач існує
	_, err := s.userService.GetUserByID(userID)
	if err != nil {
		logrus.WithError(err).Error("User not found during logout")
		return err
	}

	// TODO: Implement token blacklisting/invalidation
	// TODO: Remove user sessions from Redis/DB
	// TODO: Notify OIDC provider about logout (if required)

	logrus.WithField("user_id", userID).Info("User logged out successfully")
	return nil
}

// RefreshToken оновлює access token
func (s *authService) RefreshToken(refreshToken string) (*models.Token, error) {
	logrus.Info("AuthService: RefreshToken called")

	// Валідуємо refresh token
	refreshClaims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		logrus.WithError(err).Error("Invalid refresh token")
		return nil, err
	}

	// Отримуємо користувача з бази даних
	user, err := s.userService.GetUserByID(refreshClaims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user for refresh")
		return nil, err
	}

	// Генеруємо нові токени
	tokens, err := s.jwtService.GenerateTokens(user)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate new tokens")
		return nil, err
	}

	logrus.WithField("user_id", user.ID).Info("Tokens refreshed successfully")
	return tokens, nil
}

// GetUserInfo отримує інформацію про користувача
func (s *authService) GetUserInfo(accessToken string) (*models.User, error) {
	logrus.Info("AuthService: GetUserInfo called")

	// Валідуємо access token через JWTService
	userID, err := s.jwtService.GetUserIDFromToken(accessToken)
	if err != nil {
		logrus.WithError(err).Error("Invalid access token")
		return nil, err
	}

	// Отримуємо користувача з бази даних
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user")
		return nil, err
	}

	logrus.WithField("user_id", userID).Info("User info retrieved successfully")

	// Конвертуємо services.User в models.User
	modelUser := &models.User{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Picture:  user.Picture,
		CreateAt: user.CreatedAt,
		UpdateAt: user.UpdatedAt,
	}

	return modelUser, nil
}

// generateRandomString генерує випадковий рядок заданої довжини
