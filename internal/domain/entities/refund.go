package entities

import (
	"time"

	"github.com/google/uuid"

	"Pay2Go/internal/domain/errors"
	"Pay2Go/internal/domain/valueobjects"
)

// RefundStatus represents the state of a refund
type RefundStatus string

const (
	RefundStatusPending    RefundStatus = "pending"
	RefundStatusProcessing RefundStatus = "processing"
	RefundStatusCompleted  RefundStatus = "completed"
	RefundStatusFailed     RefundStatus = "failed"
)

// Refund represents a refund entity
type Refund struct {
	// Identity
	ID            uuid.UUID
	TransactionID uuid.UUID
	
	// Value Object
	Amount        valueobjects.Money
	
	// State
	Status        RefundStatus
	Reason        string
	
	// Provider details
	ProviderRefundID string
	
	// Error handling
	ErrorCode     string
	ErrorMessage  string
	
	// Timestamps
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ProcessedAt   *time.Time
	DeletedAt     *time.Time
}

// NewRefund creates a new refund with validation
func NewRefund(
	transactionID uuid.UUID,
	amount valueobjects.Money,
	reason string,
) (*Refund, error) {
	
	// Validate required fields
	if transactionID == uuid.Nil {
		return nil, errors.NewValidationError("transaction_id", "cannot be empty")
	}
	
	if reason == "" {
		return nil, errors.NewValidationError("reason", "cannot be empty")
	}
	
	// Validate amount
	if !amount.Currency.IsValid() {
		return nil, errors.ErrInvalidCurrency
	}
	
	now := time.Now()
	
	return &Refund{
		ID:            uuid.New(),
		TransactionID: transactionID,
		Amount:        amount,
		Reason:        reason,
		Status:        RefundStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// MarkAsProcessing transitions refund to processing state
func (r *Refund) MarkAsProcessing() error {
	if r.Status != RefundStatusPending {
		return errors.NewBusinessRuleError(
			"invalid_state_transition",
			"can only process pending refunds",
		)
	}
	
	r.Status = RefundStatusProcessing
	r.UpdatedAt = time.Now()
	return nil
}

// MarkAsCompleted marks refund as successfully completed
func (r *Refund) MarkAsCompleted(providerRefundID string) error {
	if r.Status != RefundStatusProcessing {
		return errors.NewBusinessRuleError(
			"invalid_state_transition",
			"can only complete processing refunds",
		)
	}
	
	now := time.Now()
	r.Status = RefundStatusCompleted
	r.ProviderRefundID = providerRefundID
	r.ProcessedAt = &now
	r.UpdatedAt = now
	r.ErrorCode = ""
	r.ErrorMessage = ""
	
	return nil
}

// MarkAsFailed marks refund as failed
func (r *Refund) MarkAsFailed(errorCode, errorMessage string) error {
	if r.Status != RefundStatusProcessing && r.Status != RefundStatusPending {
		return errors.NewBusinessRuleError(
			"invalid_state_transition",
			"can only fail pending or processing refunds",
		)
	}
	
	r.Status = RefundStatusFailed
	r.ErrorCode = errorCode
	r.ErrorMessage = errorMessage
	r.UpdatedAt = time.Now()
	
	return nil
}

// IsCompleted checks if refund is completed
func (r *Refund) IsCompleted() bool {
	return r.Status == RefundStatusCompleted
}

// IsFailed checks if refund has failed
func (r *Refund) IsFailed() bool {
	return r.Status == RefundStatusFailed
}

// SoftDelete marks refund as deleted
func (r *Refund) SoftDelete() {
	now := time.Now()
	r.DeletedAt = &now
	r.UpdatedAt = now
}

// IsDeleted checks if refund is soft-deleted
func (r *Refund) IsDeleted() bool {
	return r.DeletedAt != nil
}
