package usecases

import (
	"context"
	"fmt"
	"transaction-consumer/internal/domain/entities"
	"transaction-consumer/internal/domain/repositories"
	"transaction-consumer/pkg/logger"
)

type TransactionUseCase interface {
	ProcessTransaction(ctx context.Context, transaction *entities.Transaction) error
}

type transactionUseCase struct {
	transactionRepo repositories.TransactionRepository
	logger          logger.Logger
}

func NewTransactionUseCase(repo repositories.TransactionRepository, log logger.Logger) TransactionUseCase {
	return &transactionUseCase{
		transactionRepo: repo,
		logger:          log,
	}
}

func (uc *transactionUseCase) ProcessTransaction(ctx context.Context, transaction *entities.Transaction) error {
	// Validate transaction
	if !transaction.IsValid() {
		return fmt.Errorf("invalid transaction data")
	}

	exists, err := uc.transactionRepo.Exists(ctx, transaction.TransactionID)
	if err != nil {
		uc.logger.Error("Failed to check transaction existence", "error", err, "transactionID", transaction.TransactionID)
		return fmt.Errorf("failed to check transaction existence: %w", err)
	}

	if exists {
		uc.logger.Info("Transaction already exists, skipping", "transactionID", transaction.TransactionID)
		return nil
	}

	if transaction.TransactionStatus == entities.TransactionStatusFailed {
		if transaction.BalanceBefore != transaction.BalanceAfter {
			uc.logger.Warn("Failed transaction has balance change", "transactionID", transaction.TransactionID)
		}
	}

	if err := uc.transactionRepo.Create(ctx, transaction); err != nil {
		uc.logger.Error("Failed to create transaction", "error", err, "transactionID", transaction.TransactionID)
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	uc.logger.Info("Transaction processed successfully",
		"transactionID", transaction.TransactionID,
		"type", transaction.TransactionType,
		"status", transaction.TransactionStatus,
		"amount", transaction.Amount)

	return nil
}
