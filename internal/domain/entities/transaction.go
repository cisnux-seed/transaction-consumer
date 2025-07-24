package entities

import (
	"time"
)

type TransactionType string

const (
	TransactionTypeTopup    TransactionType = "TOPUP"
	TransactionTypePayment  TransactionType = "PAYMENT"
	TransactionTypeRefund   TransactionType = "REFUND"
	TransactionTypeTransfer TransactionType = "TRANSFER"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusSuccess   TransactionStatus = "SUCCESS"
	TransactionStatusFailed    TransactionStatus = "FAILED"
	TransactionStatusCancelled TransactionStatus = "CANCELLED"
)

type PaymentMethod string

type Transaction struct {
	ID                       string
	UserID                   int64
	AccountID                string
	TransactionID            string
	TransactionType          TransactionType
	TransactionStatus        TransactionStatus
	Amount                   float64
	BalanceBefore            float64
	BalanceAfter             float64
	Currency                 string
	Description              *string
	ExternalReference        *string
	PaymentMethod            *PaymentMethod
	Metadata                 *string
	IsAccessibleFromExternal bool
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

// IsValid validates the transaction entity
func (t *Transaction) IsValid() bool {
	return t.UserID > 0 &&
		t.AccountID != "" &&
		t.TransactionID != "" &&
		t.TransactionType != "" &&
		t.Amount > 0
}
