package adapter

import (
	"github.com/gofiber/fiber/v2"
	"Pay2Go/usecases"
	"strconv"
	"Pay2Go/entities"
)

type HTTPHandler struct {
	transactionUseCase usecases.TransactionUseCase
}

func NewHTTPHandler(transactionUseCase usecases.TransactionUseCase) *HTTPHandler {
	return &HTTPHandler{transactionUseCase: transactionUseCase}
}

func (h *HTTPHandler) CreateTransaction(c *fiber.Ctx) error {
	tx := new(entities.Transaction)
	if err := c.BodyParser(tx); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if err := h.transactionUseCase.CreateTransaction(tx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot create transaction"})
	}
	return c.Status(fiber.StatusCreated).JSON(tx)
}

func (h *HTTPHandler) GetTransactionByID(c *fiber.Ctx) error{
	id,err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid transaction ID"})
	}

	tx,err := h.transactionUseCase.GetTransactionByID(uint(id))

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Transaction not found"})
	}

	return c.Status(fiber.StatusOK).JSON(tx)
}

func (h *HTTPHandler) UpdateTransaction(c *fiber.Ctx) error {
	id,err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid transaction ID"})
	}

	tx,err := h.transactionUseCase.GetTransactionByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Transaction not found"})
	}

	if err := c.BodyParser(tx); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	
	if err := h.transactionUseCase.UpdateTransaction(tx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot update transaction"})
	}
	return c.Status(fiber.StatusOK).JSON(tx)
}