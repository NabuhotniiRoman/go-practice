package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HealthHandler містить handlers для health check
type HealthHandler struct{}

// NewHealthHandler створює новий HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health повертає статус здоров'я сервісу
// @Summary Health Check
// @Description Повертає статус здоров'я сервісу
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "oidc-api-server",
	})
	logrus.Info("Health check performed")
}
