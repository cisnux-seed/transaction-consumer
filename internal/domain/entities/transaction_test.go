package entities

import (
	"testing"
	"time"
)

func TestTransaction_IsValid(t *testing.T) {
	tests := []struct {
		name        string
		transaction Transaction
		expected    bool
	}{
		{
			name: "valid transaction",
			transaction: Transaction{
				UserID:          123,
				AccountID:       "account-123",
				TransactionID:   "trans-123",
				TransactionType: TransactionTypeTopup,
				Amount:          100.50,
			},
			expected: true,
		},
		{
			name: "invalid transaction - zero UserID",
			transaction: Transaction{
				UserID:          0,
				AccountID:       "account-123",
				TransactionID:   "trans-123",
				TransactionType: TransactionTypeTopup,
				Amount:          100.50,
			},
			expected: false,
		},
		{
			name: "invalid transaction - empty AccountID",
			transaction: Transaction{
				UserID:          123,
				AccountID:       "",
				TransactionID:   "trans-123",
				TransactionType: TransactionTypeTopup,
				Amount:          100.50,
			},
			expected: false,
		},
		{
			name: "invalid transaction - empty TransactionID",
			transaction: Transaction{
				UserID:          123,
				AccountID:       "account-123",
				TransactionID:   "",
				TransactionType: TransactionTypeTopup,
				Amount:          100.50,
			},
			expected: false,
		},
		{
			name: "invalid transaction - empty TransactionType",
			transaction: Transaction{
				UserID:        123,
				AccountID:     "account-123",
				TransactionID: "trans-123",
				Amount:        100.50,
			},
			expected: false,
		},
		{
			name: "invalid transaction - zero amount",
			transaction: Transaction{
				UserID:          123,
				AccountID:       "account-123",
				TransactionID:   "trans-123",
				TransactionType: TransactionTypeTopup,
				Amount:          0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.transaction.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTransactionConstants(t *testing.T) {
	// Test TransactionType constants
	if TransactionTypeTopup != "TOPUP" {
		t.Errorf("TransactionTypeTopup should be 'TOPUP', got %s", TransactionTypeTopup)
	}
	if TransactionTypePayment != "PAYMENT" {
		t.Errorf("TransactionTypePayment should be 'PAYMENT', got %s", TransactionTypePayment)
	}
	if TransactionTypeRefund != "REFUND" {
		t.Errorf("TransactionTypeRefund should be 'REFUND', got %s", TransactionTypeRefund)
	}
	if TransactionTypeTransfer != "TRANSFER" {
		t.Errorf("TransactionTypeTransfer should be 'TRANSFER', got %s", TransactionTypeTransfer)
	}

	// Test TransactionStatus constants
	if TransactionStatusPending != "PENDING" {
		t.Errorf("TransactionStatusPending should be 'PENDING', got %s", TransactionStatusPending)
	}
	if TransactionStatusSuccess != "SUCCESS" {
		t.Errorf("TransactionStatusSuccess should be 'SUCCESS', got %s", TransactionStatusSuccess)
	}
	if TransactionStatusFailed != "FAILED" {
		t.Errorf("TransactionStatusFailed should be 'FAILED', got %s", TransactionStatusFailed)
	}
	if TransactionStatusCancelled != "CANCELLED" {
		t.Errorf("TransactionStatusCancelled should be 'CANCELLED', got %s", TransactionStatusCancelled)
	}
}

func TestTransactionStruct(t *testing.T) {
	now := time.Now()
	description := "Test description"
	externalRef := "ext-ref-123"
	paymentMethod := PaymentMethod("GOPAY")
	metadata := `{"key": "value"}`

	transaction := Transaction{
		ID:                       "trans-id-123",
		UserID:                   456,
		AccountID:                "account-456",
		TransactionID:            "trans-456",
		TransactionType:          TransactionTypePayment,
		TransactionStatus:        TransactionStatusSuccess,
		Amount:                   250.75,
		BalanceBefore:            1000.00,
		BalanceAfter:             749.25,
		Currency:                 "IDR",
		Description:              &description,
		ExternalReference:        &externalRef,
		PaymentMethod:            &paymentMethod,
		Metadata:                 &metadata,
		IsAccessibleFromExternal: true,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	// Test all fields are set correctly
	if transaction.ID != "trans-id-123" {
		t.Errorf("Expected ID 'trans-id-123', got %s", transaction.ID)
	}
	if transaction.UserID != 456 {
		t.Errorf("Expected UserID 456, got %d", transaction.UserID)
	}
	if *transaction.Description != description {
		t.Errorf("Expected Description '%s', got %s", description, *transaction.Description)
	}
	if *transaction.PaymentMethod != paymentMethod {
		t.Errorf("Expected PaymentMethod '%s', got %s", paymentMethod, *transaction.PaymentMethod)
	}
}
