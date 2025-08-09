package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"
	"transaction-consumer/internal/domain/entities"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

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

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to create GORM DB: %v", err)
	}

	return gormDB, mock
}

func TestNewTransactionRepository(t *testing.T) {
	db, _ := setupTestDB(t)
	mockLog := &mockLogger{}

	repo := NewTransactionRepository(db, mockLog)
	if repo == nil {
		t.Error("NewTransactionRepository should not return nil")
	}
}

func TestTransactionRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	mockLog := &mockLogger{}
	repo := NewTransactionRepository(db, mockLog)

	transaction := &entities.Transaction{
		UserID:            123,
		AccountID:         "account-123",
		TransactionID:     "trans-123",
		TransactionType:   entities.TransactionTypeTopup,
		TransactionStatus: entities.TransactionStatusSuccess,
		Amount:            100.50,
		BalanceBefore:     1000.00,
		BalanceAfter:      1100.50,
		Currency:          "IDR",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Mock the INSERT query - use sqlmock.AnyArg() for boolean field to avoid mismatch
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "historical_transactions"`)).
		WithArgs(
			transaction.UserID,
			transaction.AccountID,
			transaction.TransactionID,
			string(transaction.TransactionType),
			string(transaction.TransactionStatus),
			transaction.Amount,
			transaction.BalanceBefore,
			transaction.BalanceAfter,
			transaction.Currency,
			nil,              // description
			nil,              // external_reference
			nil,              // payment_method
			nil,              // metadata
			sqlmock.AnyArg(), // is_accessible_external - use AnyArg to avoid mismatch
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("generated-id", time.Now(), time.Now()))
	mock.ExpectCommit()

	ctx := context.Background()
	err := repo.Create(ctx, transaction)

	if err != nil {
		t.Errorf("Create should not return error, got: %v", err)
	}

	if transaction.ID != "generated-id" {
		t.Errorf("Transaction ID should be set to generated ID, got: %s", transaction.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

// Add a separate test specifically for the IsAccessibleFromExternal field
func TestTransactionRepository_Create_WithAccessibleFlag(t *testing.T) {
	db, mock := setupTestDB(t)
	mockLog := &mockLogger{}
	repo := NewTransactionRepository(db, mockLog)

	// Test with explicitly set to true
	transaction := &entities.Transaction{
		UserID:                   123,
		AccountID:                "account-123",
		TransactionID:            "trans-accessible-true",
		TransactionType:          entities.TransactionTypeTopup,
		TransactionStatus:        entities.TransactionStatusSuccess,
		Amount:                   100.50,
		BalanceBefore:            1000.00,
		BalanceAfter:             1100.50,
		Currency:                 "IDR",
		IsAccessibleFromExternal: true, // Explicitly set to true
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "historical_transactions"`)).
		WithArgs(
			transaction.UserID,
			transaction.AccountID,
			transaction.TransactionID,
			string(transaction.TransactionType),
			string(transaction.TransactionStatus),
			transaction.Amount,
			transaction.BalanceBefore,
			transaction.BalanceAfter,
			transaction.Currency,
			nil,              // description
			nil,              // external_reference
			nil,              // payment_method
			nil,              // metadata
			true,             // is_accessible_external - explicitly true
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("generated-id-accessible", time.Now(), time.Now()))
	mock.ExpectCommit()

	ctx := context.Background()
	err := repo.Create(ctx, transaction)

	if err != nil {
		t.Errorf("Create should not return error, got: %v", err)
	}

	if transaction.ID != "generated-id-accessible" {
		t.Errorf("Transaction ID should be set to generated ID, got: %s", transaction.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

func TestTransactionRepository_Create_WithOptionalFields(t *testing.T) {
	db, mock := setupTestDB(t)
	mockLog := &mockLogger{}
	repo := NewTransactionRepository(db, mockLog)

	description := "Test transaction"
	externalRef := "ext-123"
	paymentMethod := entities.PaymentMethod("GOPAY")
	metadata := `{"key": "value"}`

	transaction := &entities.Transaction{
		UserID:                   123,
		AccountID:                "account-123",
		TransactionID:            "trans-123",
		TransactionType:          entities.TransactionTypePayment,
		TransactionStatus:        entities.TransactionStatusSuccess,
		Amount:                   100.50,
		BalanceBefore:            1000.00,
		BalanceAfter:             899.50,
		Currency:                 "IDR",
		Description:              &description,
		ExternalReference:        &externalRef,
		PaymentMethod:            &paymentMethod,
		Metadata:                 &metadata,
		IsAccessibleFromExternal: true,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "historical_transactions"`)).
		WithArgs(
			transaction.UserID,
			transaction.AccountID,
			transaction.TransactionID,
			string(transaction.TransactionType),
			string(transaction.TransactionStatus),
			transaction.Amount,
			transaction.BalanceBefore,
			transaction.BalanceAfter,
			transaction.Currency,
			description,
			externalRef,
			string(paymentMethod),
			metadata,
			true,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("generated-id", time.Now(), time.Now()))
	mock.ExpectCommit()

	ctx := context.Background()
	err := repo.Create(ctx, transaction)

	if err != nil {
		t.Errorf("Create should not return error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

func TestTransactionRepository_Create_Error(t *testing.T) {
	db, mock := setupTestDB(t)
	mockLog := &mockLogger{}
	repo := NewTransactionRepository(db, mockLog)

	transaction := &entities.Transaction{
		UserID:            123,
		AccountID:         "account-123",
		TransactionID:     "trans-123",
		TransactionType:   entities.TransactionTypeTopup,
		TransactionStatus: entities.TransactionStatusSuccess,
		Amount:            100.50,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "historical_transactions"`)).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	ctx := context.Background()
	err := repo.Create(ctx, transaction)

	if err == nil {
		t.Error("Create should return error when database operation fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

func TestTransactionRepository_GetByTransactionID_Found(t *testing.T) {
	db, mock := setupTestDB(t)
	mockLog := &mockLogger{}
	repo := NewTransactionRepository(db, mockLog)

	transactionID := "trans-123"

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "account_id", "transaction_id", "transaction_type",
		"transaction_status", "amount", "balance_before", "balance_after",
		"currency", "description", "external_reference", "payment_method",
		"metadata", "is_accessible_external", "created_at", "updated_at",
	}).AddRow(
		"id-123", 456, "account-456", transactionID, "TOPUP",
		"SUCCESS", 100.50, 1000.00, 1100.50,
		"IDR", "Test desc", "ext-ref", "GOPAY",
		`{"key": "value"}`, true, time.Now(), time.Now(),
	)

	// GORM adds ORDER BY and LIMIT to SELECT queries
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "historical_transactions" WHERE transaction_id = $1 ORDER BY "historical_transactions"."id" LIMIT $2`)).
		WithArgs(transactionID, 1).
		WillReturnRows(rows)

	ctx := context.Background()
	result, err := repo.GetByTransactionID(ctx, transactionID)

	if err != nil {
		t.Errorf("GetByTransactionID should not return error, got: %v", err)
	}

	if result == nil {
		t.Fatal("GetByTransactionID should return transaction when found")
	}

	if result.TransactionID != transactionID {
		t.Errorf("Expected transaction ID %s, got %s", transactionID, result.TransactionID)
	}

	if result.TransactionType != entities.TransactionTypeTopup {
		t.Errorf("Expected transaction type TOPUP, got %s", result.TransactionType)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

func TestTransactionRepository_GetByTransactionID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	mockLog := &mockLogger{}
	repo := NewTransactionRepository(db, mockLog)

	transactionID := "nonexistent-trans"

	// GORM adds ORDER BY and LIMIT to SELECT queries, and returns ErrRecordNotFound for First()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "historical_transactions" WHERE transaction_id = $1 ORDER BY "historical_transactions"."id" LIMIT $2`)).
		WithArgs(transactionID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	ctx := context.Background()
	result, err := repo.GetByTransactionID(ctx, transactionID)

	if err != nil {
		t.Errorf("GetByTransactionID should not return error when record not found, got: %v", err)
	}

	if result != nil {
		t.Error("GetByTransactionID should return nil when record not found")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

func TestTransactionRepository_GetByTransactionID_Error(t *testing.T) {
	db, mock := setupTestDB(t)
	mockLog := &mockLogger{}
	repo := NewTransactionRepository(db, mockLog)

	transactionID := "trans-123"

	// GORM adds ORDER BY and LIMIT to SELECT queries
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "historical_transactions" WHERE transaction_id = $1 ORDER BY "historical_transactions"."id" LIMIT $2`)).
		WithArgs(transactionID, 1).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	result, err := repo.GetByTransactionID(ctx, transactionID)

	if err == nil {
		t.Error("GetByTransactionID should return error when database operation fails")
	}

	if result != nil {
		t.Error("GetByTransactionID should return nil when error occurs")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

func TestTransactionRepository_Exists_True(t *testing.T) {
	db, mock := setupTestDB(t)
	mockLog := &mockLogger{}
	repo := NewTransactionRepository(db, mockLog)

	transactionID := "trans-123"

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "historical_transactions" WHERE transaction_id = $1`)).
		WithArgs(transactionID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	ctx := context.Background()
	exists, err := repo.Exists(ctx, transactionID)

	if err != nil {
		t.Errorf("Exists should not return error, got: %v", err)
	}

	if !exists {
		t.Error("Exists should return true when transaction exists")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

func TestTransactionRepository_Exists_False(t *testing.T) {
	db, mock := setupTestDB(t)
	mockLog := &mockLogger{}
	repo := NewTransactionRepository(db, mockLog)

	transactionID := "nonexistent-trans"

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "historical_transactions" WHERE transaction_id = $1`)).
		WithArgs(transactionID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	ctx := context.Background()
	exists, err := repo.Exists(ctx, transactionID)

	if err != nil {
		t.Errorf("Exists should not return error, got: %v", err)
	}

	if exists {
		t.Error("Exists should return false when transaction does not exist")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

func TestTransactionRepository_Exists_Error(t *testing.T) {
	db, mock := setupTestDB(t)
	mockLog := &mockLogger{}
	repo := NewTransactionRepository(db, mockLog)

	transactionID := "trans-123"

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "historical_transactions" WHERE transaction_id = $1`)).
		WithArgs(transactionID).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	exists, err := repo.Exists(ctx, transactionID)

	if err == nil {
		t.Error("Exists should return error when database operation fails")
	}

	if exists {
		t.Error("Exists should return false when error occurs")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

func TestTransactionModel_TableName(t *testing.T) {
	model := TransactionModel{}
	if model.TableName() != "historical_transactions" {
		t.Errorf("Expected table name 'historical_transactions', got %s", model.TableName())
	}
}

func TestTransactionRepository_entityToModel(t *testing.T) {
	mockLog := &mockLogger{}
	repo := &transactionRepository{logger: mockLog}

	description := "Test description"
	externalRef := "ext-ref-123"
	paymentMethod := entities.PaymentMethod("BANK_TRANSFER")
	metadata := `{"test": "data"}`

	entity := &entities.Transaction{
		ID:                       "trans-id-123",
		UserID:                   456,
		AccountID:                "account-456",
		TransactionID:            "trans-456",
		TransactionType:          entities.TransactionTypePayment,
		TransactionStatus:        entities.TransactionStatusSuccess,
		Amount:                   150.75,
		BalanceBefore:            1000.00,
		BalanceAfter:             849.25,
		Currency:                 "USD",
		Description:              &description,
		ExternalReference:        &externalRef,
		PaymentMethod:            &paymentMethod,
		Metadata:                 &metadata,
		IsAccessibleFromExternal: true,
		CreatedAt:                time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:                time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
	}

	model := repo.entityToModel(entity)

	if model.ID != entity.ID {
		t.Errorf("Expected ID %s, got %s", entity.ID, model.ID)
	}
	if model.TransactionType != string(entity.TransactionType) {
		t.Errorf("Expected type %s, got %s", entity.TransactionType, model.TransactionType)
	}
	if *model.PaymentMethod != string(*entity.PaymentMethod) {
		t.Errorf("Expected payment method %s, got %s", *entity.PaymentMethod, *model.PaymentMethod)
	}
}

func TestTransactionRepository_modelToEntity(t *testing.T) {
	mockLog := &mockLogger{}
	repo := &transactionRepository{logger: mockLog}

	description := "Test description"
	externalRef := "ext-ref-123"
	paymentMethod := "BANK_TRANSFER"
	metadata := `{"test": "data"}`

	model := &TransactionModel{
		ID:                       "trans-id-123",
		UserID:                   456,
		AccountID:                "account-456",
		TransactionID:            "trans-456",
		TransactionType:          "PAYMENT",
		TransactionStatus:        "SUCCESS",
		Amount:                   150.75,
		BalanceBefore:            1000.00,
		BalanceAfter:             849.25,
		Currency:                 "USD",
		Description:              &description,
		ExternalReference:        &externalRef,
		PaymentMethod:            &paymentMethod,
		Metadata:                 &metadata,
		IsAccessibleFromExternal: true,
		CreatedAt:                time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:                time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
	}

	entity := repo.modelToEntity(model)

	if entity.ID != model.ID {
		t.Errorf("Expected ID %s, got %s", model.ID, entity.ID)
	}
	if entity.TransactionType != entities.TransactionTypePayment {
		t.Errorf("Expected type PAYMENT, got %s", entity.TransactionType)
	}
	if string(*entity.PaymentMethod) != *model.PaymentMethod {
		t.Errorf("Expected payment method %s, got %s", *model.PaymentMethod, *entity.PaymentMethod)
	}
}

func TestTransactionRepository_entityToModel_NilOptionalFields(t *testing.T) {
	mockLog := &mockLogger{}
	repo := &transactionRepository{logger: mockLog}

	entity := &entities.Transaction{
		ID:                       "trans-id-123",
		UserID:                   456,
		AccountID:                "account-456",
		TransactionID:            "trans-456",
		TransactionType:          entities.TransactionTypeTopup,
		TransactionStatus:        entities.TransactionStatusSuccess,
		Amount:                   100.00,
		BalanceBefore:            1000.00,
		BalanceAfter:             1100.00,
		Currency:                 "IDR",
		IsAccessibleFromExternal: false,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	model := repo.entityToModel(entity)

	if model.Description != nil {
		t.Error("Description should be nil when not set in entity")
	}
	if model.ExternalReference != nil {
		t.Error("ExternalReference should be nil when not set in entity")
	}
	if model.PaymentMethod != nil {
		t.Error("PaymentMethod should be nil when not set in entity")
	}
	if model.Metadata != nil {
		t.Error("Metadata should be nil when not set in entity")
	}
}

func TestTransactionRepository_modelToEntity_NilOptionalFields(t *testing.T) {
	mockLog := &mockLogger{}
	repo := &transactionRepository{logger: mockLog}

	model := &TransactionModel{
		ID:                       "trans-id-123",
		UserID:                   456,
		AccountID:                "account-456",
		TransactionID:            "trans-456",
		TransactionType:          "TOPUP",
		TransactionStatus:        "SUCCESS",
		Amount:                   100.00,
		BalanceBefore:            1000.00,
		BalanceAfter:             1100.00,
		Currency:                 "IDR",
		IsAccessibleFromExternal: false,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	entity := repo.modelToEntity(model)

	if entity.Description != nil {
		t.Error("Description should be nil when not set in model")
	}
	if entity.ExternalReference != nil {
		t.Error("ExternalReference should be nil when not set in model")
	}
	if entity.PaymentMethod != nil {
		t.Error("PaymentMethod should be nil when not set in model")
	}
	if entity.Metadata != nil {
		t.Error("Metadata should be nil when not set in model")
	}
}
