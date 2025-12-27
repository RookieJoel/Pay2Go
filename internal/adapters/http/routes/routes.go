package routes
// Package routes configures HTTP routes
package routes










































}	transactions.Post("/:id/refund", transactionHandler.RefundTransaction)	transactions.Post("/:id/process", transactionHandler.ProcessPayment)	transactions.Get("/", transactionHandler.ListTransactions)	transactions.Get("/:id", transactionHandler.GetTransaction)	transactions.Post("/", transactionHandler.CreateTransaction)	transactions := protected.Group("/transactions")	// Transaction routes	protected.Use(middleware.RateLimiter())	protected.Use(middleware.Auth(partnerRepo))	protected := api.Group("")	// Protected routes (require authentication)	health.Get("/live", healthHandler.Live)	health.Get("/ready", healthHandler.Ready)	health.Get("/", healthHandler.Health)	health := api.Group("/health")	// Health check routes (no auth required)	api := app.Group("/api/v1")	// Public routes	app.Use(middleware.Recovery())	app.Use(middleware.Logger())	// Setup middleware) {	partnerRepo ports.PartnerRepository,	healthHandler *handlers.HealthHandler,	transactionHandler *handlers.TransactionHandler,	app *fiber.App,func SetupRoutes(// SetupRoutes configures all application routes)	"github.com/gofiber/fiber/v2"	"Pay2Go/internal/usecases/ports"	"Pay2Go/internal/adapters/http/middleware"	"Pay2Go/internal/adapters/http/handlers"import (