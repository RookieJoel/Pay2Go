package payment
// Package payment provides payment gateway implementations
package payment

import (
	"context"
	"fmt"

	"Pay2Go/internal/domain/entities"























































}	}		return NewMockPaymentGateway("manual")	default:		return NewMockPaymentGateway("adyen")	case "adyen":		return NewMockPaymentGateway("paypal")	case "paypal":		return NewMockPaymentGateway("stripe")	case "stripe":	switch provider {func NewPaymentGateway(provider string) ports.PaymentGateway {// Factory creates appropriate payment gateway based on provider}	return g.namefunc (g *MockPaymentGateway) GetProviderName() string {// GetProviderName returns the provider name}	return "completed", nil	// In production, query provider APIfunc (g *MockPaymentGateway) GetPaymentStatus(ctx context.Context, providerTransactionID string) (string, error) {// GetPaymentStatus checks payment status from provider}	return providerRefundID, nil		providerRefundID := fmt.Sprintf("mock_refund_%s_%s", g.name, refund.ID.String()[:8])	// In production, this would call Stripe/PayPal refund APIfunc (g *MockPaymentGateway) ProcessRefund(ctx context.Context, refund *entities.Refund, transaction *entities.Transaction) (string, error) {// ProcessRefund simulates refund processing}	return providerTransactionID, nil	// Simulate processing		providerTransactionID := fmt.Sprintf("mock_%s_%s", g.name, transaction.ID.String()[:8])	// For now, simulate successful payment	// In production, this would call Stripe/PayPal APIfunc (g *MockPaymentGateway) ProcessPayment(ctx context.Context, transaction *entities.Transaction) (string, error) {// ProcessPayment simulates payment processing}	return &MockPaymentGateway{name: name}func NewMockPaymentGateway(name string) ports.PaymentGateway {// NewMockPaymentGateway creates a new mock payment gateway}	name stringtype MockPaymentGateway struct {// MockPaymentGateway is a mock implementation for testing/demo)	"Pay2Go/internal/usecases/ports"