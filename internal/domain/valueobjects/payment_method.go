package valueobjects
package valueobjects

import (
	"strings"

	"Pay2Go/internal/domain/errors"
)

// PaymentMethod represents the method of payment












































































}	return err == nil	_, err := NewPaymentProvider(string(pp))func (pp PaymentProvider) IsValid() bool {// IsValid checks if payment provider is valid}	return string(pp)func (pp PaymentProvider) String() string {// String returns the string representation}	return PaymentProvider(provider), nil		}		return "", errors.NewValidationError("provider", "invalid payment provider")	if !validProviders[provider] {		}		"manual": true,		"adyen":  true,		"paypal": true,		"stripe": true,	validProviders := map[string]bool{		provider = strings.ToLower(strings.TrimSpace(provider))func NewPaymentProvider(provider string) (PaymentProvider, error) {// NewPaymentProvider validates and creates a PaymentProvider)	ProviderManual PaymentProvider = "manual"	ProviderAdyen  PaymentProvider = "adyen"	ProviderPayPal PaymentProvider = "paypal"	ProviderStripe PaymentProvider = "stripe"const (type PaymentProvider string// PaymentProvider represents external payment gateway providers}	return err == nil	_, err := NewPaymentMethod(string(pm))func (pm PaymentMethod) IsValid() bool {// IsValid checks if payment method is valid}	return string(pm)func (pm PaymentMethod) String() string {// String returns the string representation}	return PaymentMethod(method), nil		}		return "", errors.ErrInvalidPaymentMethod	if !validMethods[method] {		}		"crypto":        true,		"e_wallet":      true,		"bank_transfer": true,		"card":          true,	validMethods := map[string]bool{		method = strings.ToLower(strings.TrimSpace(method))func NewPaymentMethod(method string) (PaymentMethod, error) {// NewPaymentMethod validates and creates a PaymentMethod)	PaymentMethodCrypto       PaymentMethod = "crypto"	PaymentMethodEWallet      PaymentMethod = "e_wallet"	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"	PaymentMethodCard         PaymentMethod = "card"const (type PaymentMethod string