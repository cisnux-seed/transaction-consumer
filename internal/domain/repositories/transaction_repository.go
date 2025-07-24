package repositories

import (
	"context"
	"transaction-consumer/internal/domain/entities"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *entities.Transaction) error
	GetByTransactionID(ctx context.Context, transactionID string) (*entities.Transaction, error)
	Exists(ctx context.Context, transactionID string) (bool, error)
}
