package services

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SessionData представляє дані сесії користувача
type SessionData struct {
	SessionID string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
	State     string // OIDC state parameter
}

// SessionManager інтерфейс для управління сесіями
type SessionManager interface {
	CreateSession(userID, ipAddress, userAgent string) (*SessionData, error)
	GetSession(sessionID string) (*SessionData, error)
	UpdateSessionUser(sessionID, userID string) error
	DeleteSession(sessionID string) error
	CleanupExpiredSessions()
	GetUserSessions(userID string) ([]*SessionData, error)
}

// sessionManager реалізація SessionManager (in-memory)
type sessionManager struct {
	sessions map[string]*SessionData
	mutex    sync.RWMutex
	ttl      time.Duration
}

// NewSessionManager створює новий Session Manager
func NewSessionManager(ttl time.Duration) SessionManager {
	manager := &sessionManager{
		sessions: make(map[string]*SessionData),
		ttl:      ttl,
	}

	// Запускаємо горутину для очищення застарілих сесій
	go manager.cleanupRoutine()

	return manager
}

// CreateSession створює нову сесію
func (sm *sessionManager) CreateSession(userID, ipAddress, userAgent string) (*SessionData, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &SessionData{
		SessionID: sessionID,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(sm.ttl),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	sm.mutex.Lock()
	sm.sessions[sessionID] = session
	sm.mutex.Unlock()

	logrus.WithFields(logrus.Fields{
		"session_id": sessionID,
		"user_id":    userID,
		"ip_address": ipAddress,
		"expires_at": session.ExpiresAt,
	}).Info("Session created")

	return session, nil
}

// GetSession отримує сесію за ID
func (sm *sessionManager) GetSession(sessionID string) (*SessionData, error) {
	sm.mutex.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mutex.RUnlock()

	if !exists {
		return nil, nil // Session not found
	}

	// Перевіряємо чи не прострочена сесія
	if time.Now().After(session.ExpiresAt) {
		sm.DeleteSession(sessionID)
		return nil, nil // Session expired
	}

	return session, nil
}

// UpdateSessionUser оновлює user_id для сесії (після успішної автентифікації)
func (sm *sessionManager) UpdateSessionUser(sessionID, userID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil // Session not found
	}

	session.UserID = userID

	logrus.WithFields(logrus.Fields{
		"session_id": sessionID,
		"user_id":    userID,
	}).Info("Session updated with user ID")

	return nil
}

// DeleteSession видаляє сесію
func (sm *sessionManager) DeleteSession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		delete(sm.sessions, sessionID)
		logrus.WithFields(logrus.Fields{
			"session_id": sessionID,
			"user_id":    session.UserID,
		}).Info("Session deleted")
	}

	return nil
}

// CleanupExpiredSessions видаляє застарілі сесії
func (sm *sessionManager) CleanupExpiredSessions() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	cleaned := 0

	for sessionID, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, sessionID)
			cleaned++
		}
	}

	if cleaned > 0 {
		logrus.WithField("cleaned_count", cleaned).Info("Cleaned up expired sessions")
	}
}

// GetUserSessions повертає всі активні сесії користувача
func (sm *sessionManager) GetUserSessions(userID string) ([]*SessionData, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var userSessions []*SessionData
	now := time.Now()

	for _, session := range sm.sessions {
		if session.UserID == userID && now.Before(session.ExpiresAt) {
			userSessions = append(userSessions, session)
		}
	}

	return userSessions, nil
}

// cleanupRoutine періодично очищає застарілі сесії
func (sm *sessionManager) cleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.CleanupExpiredSessions()
	}
}

// generateSessionID генерує унікальний ID сесії
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sess_" + hex.EncodeToString(bytes), nil
}
