package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"Pay2Go/internal/domain/entities"
	"Pay2Go/internal/domain/errors"
	"Pay2Go/internal/domain/valueobjects"
	"Pay2Go/internal/usecases/ports"
)

// RefundTransactionInput represents input for refund operation
type RefundTransactionInput struct {
	TransactionID uuid.UUID
	PartnerID     uuid.UUID
	Amount        float64
	Currency      string
	Reason        string
	IPAddress     string
	UserAgent     string
}

// RefundTransactionUseCase handles the business logic for refunds
type RefundTransactionUseCase struct {
	transactionRepo ports.TransactionRepository
	refundRepo      ports.RefundRepository
	paymentGateway  ports.PaymentGateway
	notification    ports.NotificationService
	auditLogger     ports.AuditLogger
}

// NewRefundTransactionUseCase creates a new instance
func NewRefundTransactionUseCase(
	transactionRepo ports.TransactionRepository,
	refundRepo ports.RefundRepository,
	paymentGateway ports.PaymentGateway,
	notification ports.NotificationService,
	auditLogger ports.AuditLogger,
) *RefundTransactionUseCase {
	return &RefundTransactionUseCase{
		transactionRepo: transactionRepo,
		refundRepo:      refundRepo,
		paymentGateway:  paymentGateway,
		notification:    notification,
		auditLogger:     auditLogger,
	}
}

// Execute processes a refund
func (uc *RefundTransactionUseCase) Execute(ctx context.Context, input RefundTransactionInput) (*entities.Refund, error) {
	// Step 1: Retrieve transaction
	transaction, err := uc.transactionRepo.GetByID(ctx, input.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if transaction == nil {
		return nil, errors.ErrTransactionNotFound
	}

	// Step 2: Authorization - verify partner owns this transaction
	if transaction.PartnerID != input.PartnerID {
		return nil, errors.ErrUnauthorizedOperation
	}

	// Step 3: Business Rule - Check if transaction is refundable
	if !transaction.IsRefundable() {
		return nil, errors.ErrRefundNotAllowed
	}

	// Step 4: Business Rule - Check 90-day window
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	if transaction.CreatedAt.Before(ninetyDaysAgo) {
		return nil, errors.ErrRefundWindowExpired
	}

	// Step 5: Create Money value object for refund
	refundMoney, err := valueobjects.NewMoney(input.Amount, input.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid refund amount: %w", err)
	}

	// Step 6: Validate refund amount doesn't exceed transaction amount
	if refundMoney.IsGreaterThan(transaction.Amount) {
		return nil, errors.ErrRefundAmountExceeded
	}

	// Step 7: Check total refunded amount
	totalRefunded, err := uc.refundRepo.GetTotalRefundedAmount(ctx, transaction.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total refunded: %w", err)
	}

	if totalRefunded+input.Amount > transaction.Amount.Amount {
		return nil, errors.ErrRefundAmountExceeded
	}

	// Step 8: Validate currency matches
	if refundMoney.Currency != transaction.Amount.Currency {
		return nil, errors.NewValidationError("currency", "must match transaction currency")
	}

	// Step 9: Create Refund entity
	refund, err := entities.NewRefund(
		transaction.ID,
		refundMoney,
		input.Reason,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund entity: %w", err)
	}

	// Step 10: Persist refund
	if err := uc.refundRepo.Create(ctx, refund); err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	// Step 11: Mark refund as processing
	if err := refund.MarkAsProcessing(); err != nil {
		return nil, fmt.Errorf("failed to mark refund as processing: %w", err)
	}

	if err := uc.refundRepo.Update(ctx, refund); err != nil {
		return nil, fmt.Errorf("failed to update refund: %w", err)
	}

	// Step 12: Process refund through payment gateway
	providerRefundID, err := uc.paymentGateway.ProcessRefund(ctx, refund, transaction)
	if err != nil {
		// Refund failed
		_ = refund.MarkAsFailed("REFUND_FAILED", err.Error())
		_ = uc.refundRepo.Update(ctx, refund)

		return nil, fmt.Errorf("refund processing failed: %w", err)
	}

	// Step 13: Mark refund as completed
	if err := refund.MarkAsCompleted(providerRefundID); err != nil {
		return nil, fmt.Errorf("failed to mark refund as completed: %w", err)
	}

	if err := uc.refundRepo.Update(ctx, refund); err != nil {
		return nil, fmt.Errorf("failed to update refund: %w", err)
	}

	// Step 14: Update transaction status
	isFullRefund := (totalRefunded + input.Amount) >= transaction.Amount.Amount
	if err := transaction.MarkAsRefunded(!isFullRefund); err != nil {
		return nil, fmt.Errorf("failed to update transaction status: %w", err)
	}

	if err := uc.transactionRepo.Update(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	// Step 15: Log audit event
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogAction(ctx, ports.AuditAction{
			PartnerID:    input.PartnerID,
			Action:       "refund_completed",
			ResourceType: "refund",
			ResourceID:   refund.ID,
			IPAddress:    input.IPAddress,
			UserAgent:    input.UserAgent,
			Changes: map[string]interface{}{
				"transaction_id":     transaction.ID.String(),
				"amount":             input.Amount,
				"provider_refund_id": providerRefundID,
				"transaction_status": transaction.Status,
			},
		})
	}

	return refund, nil
}
