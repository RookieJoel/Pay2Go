// Package middleware contains HTTP middleware for security, logging, etc.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"Pay2Go/internal/domain/errors"
	"Pay2Go/internal/usecases/ports"
)

// AuthMiddleware validates API key and sets partner context
type AuthMiddleware struct {
	partnerRepo ports.PartnerRepository
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(partnerRepo ports.PartnerRepository) *AuthMiddleware {
	return &AuthMiddleware{
		partnerRepo: partnerRepo,
	}
}

// Handle validates API key and authenticates partner
func (m *AuthMiddleware) Handle(c *fiber.Ctx) error {
	// Get API key from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "missing authorization header",
		})
	}

	// Expected format: "Bearer <api_key>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "invalid authorization header format",
		})
	}

	apiKey := parts[1]
	if len(apiKey) < 8 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "invalid API key",
		})
	}

	// Get API key prefix (first 8 characters)
	prefix := apiKey[:8]

	// Find partner by prefix
	partner, err := m.partnerRepo.GetByAPIKeyPrefix(c.Context(), prefix)
	if err != nil || partner == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "invalid API key",
		})
	}

	// Validate API key
	if err := partner.ValidateAPIKey(apiKey); err != nil {
		if err == errors.ErrPartnerInactive {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": "partner account is inactive",
			})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "invalid API key",
		})
	}

	// Set partner ID in context
	c.Locals("partner_id", partner.ID)
	c.Locals("partner", partner)

	return c.Next()
}

// GetPartnerID retrieves partner ID from context
func GetPartnerID(c *fiber.Ctx) (uuid.UUID, error) {
	partnerID := c.Locals("partner_id")
	if partnerID == nil {
		return uuid.Nil, errors.ErrUnauthorizedOperation
	}

	if id, ok := partnerID.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, errors.ErrUnauthorizedOperation
}
