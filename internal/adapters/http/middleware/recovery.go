package middleware
package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// Recovery middleware recovers from panics
type Recovery struct{}























}	return c.Next()	}()		}			})				"message": "An unexpected error occurred",				"error":   "internal_server_error",			_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{			// Return 500 error			println("PANIC:", r)			// Log panic (in production, use proper logger)		if r := recover(); r != nil {	defer func() {func (m *Recovery) Handle(c *fiber.Ctx) error {// Handle recovers from panics and returns 500 error}	return &Recovery{}func NewRecovery() *Recovery {// NewRecovery creates a new recovery middleware