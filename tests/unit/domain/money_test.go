package entities
package valueobjects_test

import (
	"testing"

	"Pay2Go/internal/domain/valueobjects"
)

func TestNewMoney_Success(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		currency string
		wantErr  bool
	}{
		{
			name:     "Valid USD amount",
			amount:   10000, // $100.00
			currency: "USD",
			wantErr:  false,
		},
		{
			name:     "Valid EUR amount",
			amount:   50000, // €500.00
			currency: "EUR",
			wantErr:  false,
		},
		{





















































































































































































































































}	return money	}		panic(err)	if err != nil {	money, err := valueobjects.NewMoney(amount, currency)func mustNewMoney(amount int64, currency string) *valueobjects.Money {// Helper function}	}		})			}				t.Errorf("Equals() = %v, want %v", got, tt.want)			if got != tt.want {			// Assert			got := tt.money1.Equals(tt.money2)			// Act		t.Run(tt.name, func(t *testing.T) {	for _, tt := range tests {	}		},			want:   false,			money2: mustNewMoney(10000, "EUR"),			money1: mustNewMoney(10000, "USD"),			name:   "Same amount different currency",		{		},			want:   false,			money2: mustNewMoney(5000, "USD"),			money1: mustNewMoney(10000, "USD"),			name:   "Different amounts",		{		},			want:   true,			money2: mustNewMoney(10000, "USD"),			money1: mustNewMoney(10000, "USD"),			name:   "Equal amounts same currency",		{	}{		want   bool		money2 *valueobjects.Money		money1 *valueobjects.Money		name   string	tests := []struct {func TestMoney_Equals(t *testing.T) {}	}		})			}				t.Errorf("IsGreaterThan() = %v, want %v", got, tt.want)			if got != tt.want {			// Assert			got := tt.money1.IsGreaterThan(tt.money2)			// Act		t.Run(tt.name, func(t *testing.T) {	for _, tt := range tests {	}		},			want:   false,			money2: mustNewMoney(10000, "USD"),			money1: mustNewMoney(10000, "USD"),			name:   "Equal",		{		},			want:   false,			money2: mustNewMoney(10000, "USD"),			money1: mustNewMoney(5000, "USD"),			name:   "Less than",		{		},			want:   true,			money2: mustNewMoney(5000, "USD"),			money1: mustNewMoney(10000, "USD"),			name:   "Greater than",		{	}{		want   bool		money2 *valueobjects.Money		money1 *valueobjects.Money		name   string	tests := []struct {func TestMoney_IsGreaterThan(t *testing.T) {}	}		t.Error("Expected nil result for negative amount")	if result != nil {	}		t.Error("Expected error for negative result, got nil")	if err == nil {	// Assert	result, err := money1.Subtract(money2)	// Act	money2, _ := valueobjects.NewMoney(10000, "USD")	money1, _ := valueobjects.NewMoney(5000, "USD")	// Arrangefunc TestMoney_Subtract_ResultsInNegative(t *testing.T) {}	}		t.Errorf("Expected Amount 7000, got %d", result.Amount)	if result.Amount != 7000 {	}		t.Errorf("Expected no error, got %v", err)	if err != nil {	// Assert	result, err := money1.Subtract(money2)	// Act	money2, _ := valueobjects.NewMoney(3000, "USD")  // $30.00	money1, _ := valueobjects.NewMoney(10000, "USD") // $100.00	// Arrangefunc TestMoney_Subtract(t *testing.T) {}	}		t.Error("Expected nil result for different currencies")	if result != nil {	}		t.Error("Expected error for different currencies, got nil")	if err == nil {	// Assert	result, err := money1.Add(money2)	// Act	money2, _ := valueobjects.NewMoney(5000, "EUR")	money1, _ := valueobjects.NewMoney(10000, "USD")	// Arrangefunc TestMoney_Add_DifferentCurrencies(t *testing.T) {}	}		t.Errorf("Expected Amount 15000, got %d", result.Amount)	if result.Amount != 15000 {	}		t.Errorf("Expected no error, got %v", err)	if err != nil {	// Assert	result, err := money1.Add(money2)	// Act	money2, _ := valueobjects.NewMoney(5000, "USD")  // $50.00	money1, _ := valueobjects.NewMoney(10000, "USD") // $100.00	// Arrangefunc TestMoney_Add(t *testing.T) {}	}		})			}				t.Error("Expected nil money, got value")			if money != nil {			}				t.Error("Expected error, got nil")			if err == nil {			// Assert			money, err := valueobjects.NewMoney(tt.amount, tt.currency)			// Act		t.Run(tt.name, func(t *testing.T) {	for _, tt := range tests {	}		},			currency: "",			amount:   10000,			name:     "Empty currency",		{		},			currency: "XXX",			amount:   10000,			name:     "Invalid currency",		{		},			currency: "USD",			amount:   10000001, // $100,000.01			name:     "Amount exceeds maximum",		{		},			currency: "USD",			amount:   0,			name:     "Zero amount",		{		},			currency: "USD",			amount:   -10000,			name:     "Negative amount",		{	}{		currency string		amount   int64		name     string	tests := []struct {func TestNewMoney_Invalid(t *testing.T) {}	}		})			}				}					t.Errorf("Currency = %s, want %s", money.Currency, tt.currency)				if money.Currency != tt.currency {				}					t.Errorf("Amount = %d, want %d", money.Amount, tt.amount)				if money.Amount != tt.amount {			if !tt.wantErr {			}				return				t.Errorf("NewMoney() error = %v, wantErr %v", err, tt.wantErr)			if (err != nil) != tt.wantErr {			// Assert			money, err := valueobjects.NewMoney(tt.amount, tt.currency)			// Act		t.Run(tt.name, func(t *testing.T) {	for _, tt := range tests {	}		},			wantErr:  false,			currency: "USD",			amount:   10000000, // $100,000.00			name:     "Maximum valid amount",		{		},			wantErr:  false,			currency: "USD",			amount:   1, // $0.01			name:     "Minimum valid amount",