package models

import (
	"errors"
	"fmt"
)

// TransactionStatus type for status of transaction, defined by the payment provided
type TransactionStatus string

// PaymentProvider type for payment provider used to process transaction
type PaymentProvider string

// TransactionType type to handle type of transaction
type TransactionType string

var (
	// TransactionStatusSucceeded status for succeeded transaction
	TransactionStatusSucceeded TransactionStatus = "succeeded"
	// TransactionStatusFailure status for failed transaction
	TransactionStatusFailure TransactionStatus = "failure"
	// TransactionStatusPending status for pending transaction
	TransactionStatusPending TransactionStatus = "pending"

	// PaymentProviderStripe represents the Stripe integration
	PaymentProviderStripe PaymentProvider = "stripe"
	// PaymentProviderMock represents a mock integration
	PaymentProviderMock PaymentProvider = "mock"

	// TransactionTypeCharge type for processed transactions
	TransactionTypeCharge TransactionType = "charge"
	// TransactionTypeRefund type for refunded transactions
	TransactionTypeRefund TransactionType = "refund"
)

// Transaction struct to process and store a transaction
type Transaction struct {
	TransactionID    string                 `json:"transaction_id"`
	Status           TransactionStatus      `json:"status"`
	Description      string                 `json:"description"`
	FailureReason    string                 `json:"failure_reason,omitempty"`
	Provider         PaymentProvider        `json:"payment_provider"`
	Amount           int                    `json:"amount"`
	Currency         string                 `json:"currency"`
	Type             TransactionType        `json:"type"`
	AdditionalFields map[string]interface{} `json:"additional_fields"`
}

// TransactionInput inputs to perform a transaction
type TransactionInput struct {
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	PaymentMethod string `json:"payment_method"`
	Description   string `json:"description"`
}

var (
	// ErrInvalidAmount error when amount is equal or less than zero
	ErrInvalidAmount = errors.New("invalid amount")
	// ErrMissingCurrency error when currency is missing
	ErrMissingCurrency = errors.New("missing currency")
	// ErrMissingPaymentMethod error when payment method is missing
	ErrMissingPaymentMethod = errors.New("missing payment method")
)

// Validate validate the inputs required for a transaction
func (ti *TransactionInput) Validate() error {
	if ti.Amount <= 0 {
		return ErrInvalidAmount
	}

	if ti.Currency == "" {
		return ErrMissingCurrency
	}

	if ti.PaymentMethod == "" {
		return ErrMissingPaymentMethod
	}

	if ti.Description == "" {
		ti.Description = fmt.Sprintf("Transaction for payment amount of %d", ti.Amount)
	}

	return nil
}
