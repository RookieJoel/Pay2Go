package errors
// Package errors defines domain-specific errors
// These are business rule violations, not technical errors
package errors

import (
	"errors"
	"fmt"
)





































































}	}		Message: fmt.Sprintf("%s: %s", rule, message),		Code:    "BUSINESS_RULE_VIOLATION",	return &DomainError{func NewBusinessRuleError(rule, message string) *DomainError {// Business rule errors}	}		Message: fmt.Sprintf("%s: %s", field, message),		Code:    "VALIDATION_ERROR",	return &DomainError{func NewValidationError(field, message string) *DomainError {// Validation errors}	}		Err:     err,		Message: message,		Code:    code,	return &DomainError{func NewDomainError(code, message string, err error) *DomainError {// NewDomainError creates a new domain error}	return e.Errfunc (e *DomainError) Unwrap() error {}	return fmt.Sprintf("[%s] %s", e.Code, e.Message)	}		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)	if e.Err != nil {func (e *DomainError) Error() string {}	Err     error	Message string	Code    stringtype DomainError struct {// DomainError represents a domain-specific error with context)	ErrUnauthorizedOperation  = errors.New("unauthorized operation")	ErrAmountAboveMaximum     = errors.New("amount above maximum allowed")	ErrAmountBelowMinimum     = errors.New("amount below minimum allowed")	// Business rule errors		ErrRefundWindowExpired    = errors.New("refund window has expired")	ErrRefundNotAllowed       = errors.New("refund not allowed for this transaction")	ErrRefundAmountExceeded   = errors.New("refund amount exceeds transaction amount")	// Refund errors		ErrInvalidAPIKey          = errors.New("invalid API key")	ErrPartnerInactive        = errors.New("partner is inactive")	ErrPartnerNotFound        = errors.New("partner not found")	// Partner errors		ErrDuplicateTransaction   = errors.New("duplicate transaction detected")	ErrTransactionNotFound    = errors.New("transaction not found")	ErrInvalidStatus          = errors.New("invalid transaction status")	ErrInvalidPaymentMethod   = errors.New("invalid payment method")	ErrInvalidCurrency        = errors.New("invalid currency code")	ErrInvalidAmount          = errors.New("invalid transaction amount")	// Transaction errorsvar (// Common domain errors