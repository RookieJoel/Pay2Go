package handlers
package handlers

import (








































}	})		"status": "alive",	return c.JSON(fiber.Map{func (h *HealthHandler) Live(c *fiber.Ctx) error {// Live handles GET /health/live}	})		"status": "ready",	return c.JSON(fiber.Map{	// In production, check database connectivity, external services, etc.func (h *HealthHandler) Ready(c *fiber.Ctx) error {// Ready handles GET /health/ready}	return c.JSON(response)	}		Version:   "1.0.0",		Timestamp: time.Now(),		Status:    "healthy",	response := dto.HealthCheckResponse{func (h *HealthHandler) Check(c *fiber.Ctx) error {// Check handles GET /health}	return &HealthHandler{}func NewHealthHandler() *HealthHandler {// NewHealthHandler creates a new health handlertype HealthHandler struct{}// HealthHandler handles health check requests)	"Pay2Go/internal/adapters/http/dto"	"github.com/gofiber/fiber/v2"	"time"