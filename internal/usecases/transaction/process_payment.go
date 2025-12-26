package transaction

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"Pay2Go/internal/domain/entities"
	"Pay2Go/internal/domain/errors"
	"Pay2Go/internal/usecases/ports"
)

// ProcessPaymentUseCase handles the business logic for processing payments
// This orchestrates the interaction with external payment providers
type ProcessPaymentUseCase struct {
	transactionRepo ports.TransactionRepository
	paymentGateway  ports.PaymentGateway
	notification    ports.NotificationService
	auditLogger     ports.AuditLogger
}

// NewProcessPaymentUseCase creates a new instance
func NewProcessPaymentUseCase(
	transactionRepo ports.TransactionRepository,
	paymentGateway ports.PaymentGateway,
	notification ports.NotificationService,
	auditLogger ports.AuditLogger,
) *ProcessPaymentUseCase {
	return &ProcessPaymentUseCase{
		transactionRepo: transactionRepo,
		paymentGateway:  paymentGateway,
		notification:    notification,
		auditLogger:     auditLogger,
	}
}

// Execute processes a payment through the payment gateway
func (uc *ProcessPaymentUseCase) Execute(ctx context.Context, transactionID uuid.UUID) error {
	// Step 1: Retrieve transaction
	transaction, err := uc.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	
	if transaction == nil {
		return errors.ErrTransactionNotFound
	}
	
	// Step 2: Validate transaction state
	if !transaction.IsPending() && !transaction.IsFailed() {
		return errors.NewBusinessRuleError(
			"invalid_state",
			fmt.Sprintf("transaction is in %s state, cannot process", transaction.Status),
		)
	}
	
	// Step 3: Mark as processing
	if err := transaction.MarkAsProcessing(); err != nil {
		return fmt.Errorf("failed to mark as processing: %w", err)
	}
	
	if err := uc.transactionRepo.Update(ctx, transaction); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	
	// Step 4: Process payment through gateway
	providerTxnID, err := uc.paymentGateway.ProcessPayment(ctx, transaction)
	if err != nil {
		// Payment failed - mark transaction as failed
		_ = transaction.MarkAsFailed("PAYMENT_FAILED", err.Error())
		_ = uc.transactionRepo.Update(ctx, transaction)
		
		// Log audit event
		if uc.auditLogger != nil {
			_ = uc.auditLogger.LogAction(ctx, ports.AuditAction{
				PartnerID:    transaction.PartnerID,
				Action:       "payment_failed",
				ResourceType: "transaction",
				ResourceID:   transaction.ID,
				Changes: map[string]interface{}{
					"error":  err.Error(),
					"status": transaction.Status,
				},
			})
		}
		
		return fmt.Errorf("payment processing failed: %w", err)
	}
	
	// Step 5: Mark transaction as completed
	if err := transaction.MarkAsCompleted(providerTxnID); err != nil {
		return fmt.Errorf("failed to mark as completed: %w", err)
	}
	
	if err := uc.transactionRepo.Update(ctx, transaction); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	
	// Step 6: Log audit event
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogAction(ctx, ports.AuditAction{
			PartnerID:    transaction.PartnerID,
			Action:       "payment_completed",
			ResourceType: "transaction",
			ResourceID:   transaction.ID,
			Changes: map[string]interface{}{
				"provider_transaction_id": providerTxnID,
				"status":                   transaction.Status,
			},
		})
	}
	
	// Step 7: Send webhook notification (async, fire-and-forget)
	// In production, this would be done via message queue
	if uc.notification != nil {
		go func() {
			payload := map[string]interface{}{
				"event":         "payment.completed",
				"transaction_id": transaction.ID.String(),
				"status":        transaction.Status,
				"amount":        transaction.Amount.Amount,
				"currency":      transaction.Amount.Currency,
			}
			// Note: Webhook URL would come from partner configuration
			// _ = uc.notification.SendWebhook(context.Background(), webhookURL, payload)
			_ = payload
		}()
	}
	
	return nil
}

// RetryFailedPaymentUseCase handles retrying failed payments
type RetryFailedPaymentUseCase struct {
	transactionRepo ports.TransactionRepository
	paymentGateway  ports.PaymentGateway
	auditLogger     ports.AuditLogger
}

// NewRetryFailedPaymentUseCase creates a new instance
func NewRetryFailedPaymentUseCase(
	transactionRepo ports.TransactionRepository,
	paymentGateway ports.PaymentGateway,
	auditLogger ports.AuditLogger,
) *RetryFailedPaymentUseCase {
	return &RetryFailedPaymentUseCase{
		transactionRepo: transactionRepo,
		paymentGateway:  paymentGateway,
		auditLogger:     auditLogger,
	}
}

// Execute retries a failed payment
func (uc *RetryFailedPaymentUseCase) Execute(ctx context.Context, transactionID uuid.UUID) error {
	// Get transaction
	transaction, err := uc.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	
	if transaction == nil {
		return errors.ErrTransactionNotFound
	}
	
	// Check if retry is allowed (business rule: max 3 retries)
	if !transaction.CanRetry() {
		return errors.NewBusinessRuleError(
			"max_retries_exceeded",
			"transaction has reached maximum retry attempts",
		)
	}
	
	// Increment retry count
	if err := transaction.IncrementRetryCount(); err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}
	
	if err := uc.transactionRepo.Update(ctx, transaction); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	
	// Log audit event
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogAction(ctx, ports.AuditAction{
			PartnerID:    transaction.PartnerID,
			Action:       "payment_retry",
			ResourceType: "transaction",
			ResourceID:   transaction.ID,
			Changes: map[string]interface{}{
				"retry_count": transaction.RetryCount,
			},
		})
	}
	
	// Process payment again
	processUseCase := NewProcessPaymentUseCase(
		uc.transactionRepo,
		uc.paymentGateway,
		nil,
		uc.auditLogger,
	)
	
	return processUseCase.Execute(ctx, transactionID)
}
