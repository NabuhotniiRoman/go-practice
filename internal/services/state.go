package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// StateService інтерфейс для роботи з CSRF state параметрами
type StateService interface {
	GenerateState(sessionID string) (string, error)
	ValidateState(state string) (string, error)
	CleanupExpiredStates()
}

// stateEntry представляє запис state в пам'яті
type stateEntry struct {
	SessionID string
	ExpiresAt time.Time
}

// stateService реалізація StateService
type stateService struct {
	states map[string]*stateEntry
	mutex  sync.RWMutex
	ttl    time.Duration
}

// NewStateService створює новий State сервіс
func NewStateService(ttl time.Duration) StateService {
	service := &stateService{
		states: make(map[string]*stateEntry),
		ttl:    ttl,
	}

	// Запускаємо горутину для очищення застарілих state
	go service.cleanupRoutine()

	return service
}

// GenerateState генерує новий state параметр для CSRF захисту
func (s *stateService) GenerateState(sessionID string) (string, error) {
	// Генеруємо криптографічно стійкий випадковий state
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}

	state := hex.EncodeToString(randomBytes)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Зберігаємо state з TTL
	s.states[state] = &stateEntry{
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(s.ttl),
	}

	logrus.WithFields(logrus.Fields{
		"state":      state[:10] + "...",
		"session_id": sessionID,
		"expires_at": s.states[state].ExpiresAt,
	}).Debug("Generated new state parameter")

	return state, nil
}

// ValidateState валідує state параметр і повертає session_id
func (s *stateService) ValidateState(state string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	entry, exists := s.states[state]
	if !exists {
		return "", fmt.Errorf("invalid state parameter")
	}

	// Перевіряємо, чи не закінчився TTL
	if time.Now().After(entry.ExpiresAt) {
		delete(s.states, state)
		return "", fmt.Errorf("state parameter expired")
	}

	sessionID := entry.SessionID

	// Видаляємо state після використання (одноразове використання)
	delete(s.states, state)

	logrus.WithFields(logrus.Fields{
		"state":      state[:10] + "...",
		"session_id": sessionID,
	}).Debug("State parameter validated successfully")

	return sessionID, nil
}

// CleanupExpiredStates видаляє застарілі state параметри
func (s *stateService) CleanupExpiredStates() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	cleaned := 0

	for state, entry := range s.states {
		if now.After(entry.ExpiresAt) {
			delete(s.states, state)
			cleaned++
		}
	}

	if cleaned > 0 {
		logrus.WithField("cleaned_count", cleaned).Debug("Cleaned up expired state parameters")
	}
}

// cleanupRoutine періодично очищує застарілі state параметри
func (s *stateService) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.CleanupExpiredStates()
	}
}
