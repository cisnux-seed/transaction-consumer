package usecases

import (
	"context"
	"errors"
	"testing"
	"transaction-consumer/internal/domain/entities"
	_ "transaction-consumer/pkg/logger"
)

// Mock repository for testing
type mockTransactionRepository struct {
	transactions map[string]*entities.Transaction
	createError  error
	existsError  error
}

func (m *mockTransactionRepository) Create(ctx context.Context, transaction *entities.Transaction) error {
	if m.createError != nil {
		return m.createError
	}
	if m.transactions == nil {
		m.transactions = make(map[string]*entities.Transaction)
	}
	m.transactions[transaction.TransactionID] = transaction
	return nil
}

func (m *mockTransactionRepository) GetByTransactionID(ctx context.Context, transactionID string) (*entities.Transaction, error) {
	if m.transactions == nil {
		return nil, nil
	}
	transaction, exists := m.transactions[transactionID]
	if !exists {
		return nil, nil
	}
	return transaction, nil
}

func (m *mockTransactionRepository) Exists(ctx context.Context, transactionID string) (bool, error) {
	if m.existsError != nil {
		return false, m.existsError
	}
	if m.transactions == nil {
		return false, nil
	}
	_, exists := m.transactions[transactionID]
	return exists, nil
}

// Mock logger for testing
type mockLogger struct {
	debugMsgs []string
	infoMsgs  []string
	warnMsgs  []string
	errorMsgs []string
}

func (m *mockLogger) Debug(msg string, args ...interface{}) {
	if m.debugMsgs == nil {
		m.debugMsgs = []string{}
	}
	m.debugMsgs = append(m.debugMsgs, msg)
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	if m.infoMsgs == nil {
		m.infoMsgs = []string{}
	}
	m.infoMsgs = append(m.infoMsgs, msg)
}

func (m *mockLogger) Warn(msg string, args ...interface{}) {
	if m.warnMsgs == nil {
		m.warnMsgs = []string{}
	}
	m.warnMsgs = append(m.warnMsgs, msg)
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	if m.errorMsgs == nil {
		m.errorMsgs = []string{}
	}
	m.errorMsgs = append(m.errorMsgs, msg)
}

func (m *mockLogger) Fatal(msg string, args ...interface{}) {
	m.Error(msg, args...)
}

func TestNewTransactionUseCase(t *testing.T) {
	mockRepo := &mockTransactionRepository{}
	mockLog := &mockLogger{}

	useCase := NewTransactionUseCase(mockRepo, mockLog)
	if useCase == nil {
		t.Error("NewTransactionUseCase should not return nil")
	}
}

func TestTransactionUseCase_ProcessTransaction_Success(t *testing.T) {
	mockRepo := &mockTransactionRepository{}
	mockLog := &mockLogger{}
	useCase := NewTransactionUseCase(mockRepo, mockLog)

	transaction := &entities.Transaction{
		UserID:            123,
		AccountID:         "account-123",
		TransactionID:     "trans-123",
		TransactionType:   entities.TransactionTypeTopup,
		TransactionStatus: entities.TransactionStatusSuccess,
		Amount:            100.50,
		BalanceBefore:     1000.00,
		BalanceAfter:      1100.50,
	}

	ctx := context.Background()
	err := useCase.ProcessTransaction(ctx, transaction)

	if err != nil {
		t.Errorf("ProcessTransaction should not return error, got: %v", err)
	}

	// Check if transaction was stored
	exists, _ := mockRepo.Exists(ctx, transaction.TransactionID)
	if !exists {
		t.Error("Transaction should exist in repository after processing")
	}

	// Check if success was logged
	found := false
	for _, msg := range mockLog.infoMsgs {
		if msg == "Transaction processed successfully" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Success message should be logged")
	}
}

func TestTransactionUseCase_ProcessTransaction_InvalidTransaction(t *testing.T) {
	mockRepo := &mockTransactionRepository{}
	mockLog := &mockLogger{}
	useCase := NewTransactionUseCase(mockRepo, mockLog)

	// Invalid transaction (missing required fields)
	transaction := &entities.Transaction{
		TransactionID: "trans-123",
	}

	ctx := context.Background()
	err := useCase.ProcessTransaction(ctx, transaction)

	if err == nil {
		t.Error("ProcessTransaction should return error for invalid transaction")
	}

	if err.Error() != "invalid transaction data" {
		t.Errorf("Expected 'invalid transaction data' error, got: %v", err)
	}
}

func TestTransactionUseCase_ProcessTransaction_ExistsError(t *testing.T) {
	mockRepo := &mockTransactionRepository{
		existsError: errors.New("database error"),
	}
	mockLog := &mockLogger{}
	useCase := NewTransactionUseCase(mockRepo, mockLog)

	transaction := &entities.Transaction{
		UserID:            123,
		AccountID:         "account-123",
		TransactionID:     "trans-123",
		TransactionType:   entities.TransactionTypeTopup,
		TransactionStatus: entities.TransactionStatusSuccess,
		Amount:            100.50,
	}

	ctx := context.Background()
	err := useCase.ProcessTransaction(ctx, transaction)

	if err == nil {
		t.Error("ProcessTransaction should return error when repository.Exists fails")
	}

	// Check if error was logged
	found := false
	for _, msg := range mockLog.errorMsgs {
		if msg == "Failed to check transaction existence" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Error message should be logged")
	}
}

func TestTransactionUseCase_ProcessTransaction_AlreadyExists(t *testing.T) {
	mockRepo := &mockTransactionRepository{
		transactions: map[string]*entities.Transaction{
			"existing-trans": {TransactionID: "existing-trans"},
		},
	}
	mockLog := &mockLogger{}
	useCase := NewTransactionUseCase(mockRepo, mockLog)

	transaction := &entities.Transaction{
		UserID:            123,
		AccountID:         "account-123",
		TransactionID:     "existing-trans",
		TransactionType:   entities.TransactionTypeTopup,
		TransactionStatus: entities.TransactionStatusSuccess,
		Amount:            100.50,
	}

	ctx := context.Background()
	err := useCase.ProcessTransaction(ctx, transaction)

	if err != nil {
		t.Errorf("ProcessTransaction should not return error for existing transaction, got: %v", err)
	}

	// Check if skip message was logged
	found := false
	for _, msg := range mockLog.infoMsgs {
		if msg == "Transaction already exists, skipping" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Skip message should be logged")
	}
}

func TestTransactionUseCase_ProcessTransaction_CreateError(t *testing.T) {
	mockRepo := &mockTransactionRepository{
		createError: errors.New("create error"),
	}
	mockLog := &mockLogger{}
	useCase := NewTransactionUseCase(mockRepo, mockLog)

	transaction := &entities.Transaction{
		UserID:            123,
		AccountID:         "account-123",
		TransactionID:     "trans-123",
		TransactionType:   entities.TransactionTypeTopup,
		TransactionStatus: entities.TransactionStatusSuccess,
		Amount:            100.50,
	}

	ctx := context.Background()
	err := useCase.ProcessTransaction(ctx, transaction)

	if err == nil {
		t.Error("ProcessTransaction should return error when repository.Create fails")
	}

	// Check if error was logged
	found := false
	for _, msg := range mockLog.errorMsgs {
		if msg == "Failed to create transaction" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Create error message should be logged")
	}
}

func TestTransactionUseCase_ProcessTransaction_FailedTransactionWithBalanceChange(t *testing.T) {
	mockRepo := &mockTransactionRepository{}
	mockLog := &mockLogger{}
	useCase := NewTransactionUseCase(mockRepo, mockLog)

	// Failed transaction with balance change (suspicious)
	transaction := &entities.Transaction{
		UserID:            123,
		AccountID:         "account-123",
		TransactionID:     "trans-failed",
		TransactionType:   entities.TransactionTypePayment,
		TransactionStatus: entities.TransactionStatusFailed,
		Amount:            100.50,
		BalanceBefore:     1000.00,
		BalanceAfter:      900.00, // Balance changed even though transaction failed
	}

	ctx := context.Background()
	err := useCase.ProcessTransaction(ctx, transaction)

	if err != nil {
		t.Errorf("ProcessTransaction should not return error, got: %v", err)
	}

	// Check if warning was logged
	found := false
	for _, msg := range mockLog.warnMsgs {
		if msg == "Failed transaction has balance change" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Warning message should be logged for failed transaction with balance change")
	}
}

func TestTransactionUseCase_ProcessTransaction_FailedTransactionNoBalanceChange(t *testing.T) {
	mockRepo := &mockTransactionRepository{}
	mockLog := &mockLogger{}
	useCase := NewTransactionUseCase(mockRepo, mockLog)

	// Failed transaction without balance change (expected)
	transaction := &entities.Transaction{
		UserID:            123,
		AccountID:         "account-123",
		TransactionID:     "trans-failed-ok",
		TransactionType:   entities.TransactionTypePayment,
		TransactionStatus: entities.TransactionStatusFailed,
		Amount:            100.50,
		BalanceBefore:     1000.00,
		BalanceAfter:      1000.00, // No balance change for failed transaction
	}

	ctx := context.Background()
	err := useCase.ProcessTransaction(ctx, transaction)

	if err != nil {
		t.Errorf("ProcessTransaction should not return error, got: %v", err)
	}

	// Check that no warning was logged
	for _, msg := range mockLog.warnMsgs {
		if msg == "Failed transaction has balance change" {
			t.Error("Warning should not be logged for failed transaction without balance change")
		}
	}
}

func TestTransactionUseCase_ProcessTransaction_AllTransactionTypes(t *testing.T) {
	mockRepo := &mockTransactionRepository{}
	mockLog := &mockLogger{}
	useCase := NewTransactionUseCase(mockRepo, mockLog)

	transactionTypes := []entities.TransactionType{
		entities.TransactionTypeTopup,
		entities.TransactionTypePayment,
		entities.TransactionTypeRefund,
		entities.TransactionTypeTransfer,
	}

	ctx := context.Background()

	for _, transactionType := range transactionTypes {
		transaction := &entities.Transaction{
			UserID:            123,
			AccountID:         "account-123",
			TransactionID:     "trans-" + string(transactionType),
			TransactionType:   transactionType,
			TransactionStatus: entities.TransactionStatusSuccess,
			Amount:            100.50,
			BalanceBefore:     1000.00,
			BalanceAfter:      1100.50,
		}

		err := useCase.ProcessTransaction(ctx, transaction)
		if err != nil {
			t.Errorf("ProcessTransaction should not return error for %s, got: %v", transactionType, err)
		}

		// Verify transaction was stored
		exists, _ := mockRepo.Exists(ctx, transaction.TransactionID)
		if !exists {
			t.Errorf("Transaction %s should exist in repository after processing", transactionType)
		}
	}

	// Check that we have the right number of success messages
	successCount := 0
	for _, msg := range mockLog.infoMsgs {
		if msg == "Transaction processed successfully" {
			successCount++
		}
	}
	if successCount != len(transactionTypes) {
		t.Errorf("Expected %d success messages, got %d", len(transactionTypes), successCount)
	}
}
