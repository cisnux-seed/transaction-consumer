package deliveries

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
	"transaction-consumer/internal/domain/entities"
)

// Mock use case for testing
type mockTransactionUseCase struct {
	processError error
	processed    []*entities.Transaction
}

func (m *mockTransactionUseCase) ProcessTransaction(ctx context.Context, transaction *entities.Transaction) error {
	if m.processError != nil {
		return m.processError
	}
	if m.processed == nil {
		m.processed = []*entities.Transaction{}
	}
	m.processed = append(m.processed, transaction)
	return nil
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

func TestNewTransactionHandler(t *testing.T) {
	mockUseCase := &mockTransactionUseCase{}
	mockLog := &mockLogger{}

	handler := NewTransactionHandler(mockUseCase, mockLog)
	if handler == nil {
		t.Error("NewTransactionHandler should not return nil")
	}
}

func TestTransactionHandler_HandleMessage_Success(t *testing.T) {
	mockUseCase := &mockTransactionUseCase{}
	mockLog := &mockLogger{}
	handler := NewTransactionHandler(mockUseCase, mockLog)

	// Create a valid Kafka message
	kafkaMsg := KafkaTransactionMessage{
		ID:                       "trans-id-123",
		UserID:                   456,
		AccountID:                "account-456",
		TransactionID:            "trans-456",
		TransactionType:          "TOPUP",
		TransactionStatus:        "SUCCESS",
		Amount:                   250.75,
		BalanceBefore:            1000.00,
		BalanceAfter:             1250.75,
		Currency:                 "IDR",
		Description:              "Test transaction",
		PaymentMethod:            "GOPAY",
		IsAccessibleFromExternal: true,
		CreatedAt:                []interface{}{2024.0, 1.0, 15.0, 10.0, 30.0, 45.0, 0.0},
		UpdatedAt:                []interface{}{2024.0, 1.0, 15.0, 10.0, 30.0, 45.0, 0.0},
	}

	message, err := json.Marshal(kafkaMsg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	ctx := context.Background()
	err = handler.HandleMessage(ctx, message)

	if err != nil {
		t.Errorf("HandleMessage should not return error, got: %v", err)
	}

	// Check if transaction was processed
	if len(mockUseCase.processed) != 1 {
		t.Errorf("Expected 1 processed transaction, got %d", len(mockUseCase.processed))
	}

	processedTx := mockUseCase.processed[0]
	if processedTx.TransactionID != "trans-456" {
		t.Errorf("Expected transaction ID 'trans-456', got %s", processedTx.TransactionID)
	}
	if processedTx.TransactionType != entities.TransactionTypeTopup {
		t.Errorf("Expected transaction type TOPUP, got %s", processedTx.TransactionType)
	}
}

func TestTransactionHandler_HandleMessage_InvalidJSON(t *testing.T) {
	mockUseCase := &mockTransactionUseCase{}
	mockLog := &mockLogger{}
	handler := NewTransactionHandler(mockUseCase, mockLog)

	invalidJSON := []byte(`{"invalid": json}`)

	ctx := context.Background()
	err := handler.HandleMessage(ctx, invalidJSON)

	if err == nil {
		t.Error("HandleMessage should return error for invalid JSON")
	}

	if len(mockUseCase.processed) != 0 {
		t.Error("No transaction should be processed for invalid JSON")
	}
}

func TestTransactionHandler_HandleMessage_ProcessError(t *testing.T) {
	mockUseCase := &mockTransactionUseCase{
		processError: errors.New("process error"),
	}
	mockLog := &mockLogger{}
	handler := NewTransactionHandler(mockUseCase, mockLog)

	kafkaMsg := KafkaTransactionMessage{
		ID:                "trans-id-123",
		UserID:            456,
		AccountID:         "account-456",
		TransactionID:     "trans-456",
		TransactionType:   "TOPUP",
		TransactionStatus: "SUCCESS",
		Amount:            250.75,
		CreatedAt:         []interface{}{2024.0, 1.0, 15.0, 10.0, 30.0, 45.0},
		UpdatedAt:         []interface{}{2024.0, 1.0, 15.0, 10.0, 30.0, 45.0},
	}

	message, _ := json.Marshal(kafkaMsg)

	ctx := context.Background()
	err := handler.HandleMessage(ctx, message)

	if err == nil {
		t.Error("HandleMessage should return error when use case fails")
	}
}

func TestTransactionHandler_parseTimestamp_Valid(t *testing.T) {
	mockUseCase := &mockTransactionUseCase{}
	mockLog := &mockLogger{}
	handler := NewTransactionHandler(mockUseCase, mockLog)

	timestampArray := []interface{}{2024.0, 1.0, 15.0, 10.0, 30.0, 45.0, 500000000.0}

	result, err := handler.parseTimestamp(timestampArray)
	if err != nil {
		t.Errorf("parseTimestamp should not return error, got: %v", err)
	}

	expected := time.Date(2024, 1, 15, 10, 30, 45, 500000000, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestTransactionHandler_parseTimestamp_ValidWithoutNanoseconds(t *testing.T) {
	mockUseCase := &mockTransactionUseCase{}
	mockLog := &mockLogger{}
	handler := NewTransactionHandler(mockUseCase, mockLog)

	timestampArray := []interface{}{2024.0, 1.0, 15.0, 10.0, 30.0, 45.0}

	result, err := handler.parseTimestamp(timestampArray)
	if err != nil {
		t.Errorf("parseTimestamp should not return error, got: %v", err)
	}

	expected := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestTransactionHandler_parseTimestamp_Invalid(t *testing.T) {
	mockUseCase := &mockTransactionUseCase{}
	mockLog := &mockLogger{}
	handler := NewTransactionHandler(mockUseCase, mockLog)

	timestampArray := []interface{}{2024.0, 1.0} // Too few elements

	_, err := handler.parseTimestamp(timestampArray)
	if err == nil {
		t.Error("parseTimestamp should return error for invalid timestamp array")
	}
}

func TestTransactionHandler_kafkaMessageToEntity_Success(t *testing.T) {
	mockUseCase := &mockTransactionUseCase{}
	mockLog := &mockLogger{}
	handler := NewTransactionHandler(mockUseCase, mockLog)

	externalRef := "ext-ref-123"
	metadata := `{"key": "value"}`

	kafkaMsg := &KafkaTransactionMessage{
		ID:                       "trans-id-123",
		UserID:                   456,
		AccountID:                "account-456",
		TransactionID:            "trans-456",
		TransactionType:          "PAYMENT",
		TransactionStatus:        "SUCCESS",
		Amount:                   150.25,
		BalanceBefore:            1000.00,
		BalanceAfter:             849.75,
		Currency:                 "IDR",
		Description:              "Test payment",
		ExternalReference:        &externalRef,
		PaymentMethod:            "BANK_TRANSFER",
		Metadata:                 &metadata,
		IsAccessibleFromExternal: true,
		CreatedAt:                []interface{}{2024.0, 2.0, 20.0, 14.0, 15.0, 30.0},
		UpdatedAt:                []interface{}{2024.0, 2.0, 20.0, 14.0, 15.0, 30.0},
	}

	result, err := handler.kafkaMessageToEntity(kafkaMsg)
	if err != nil {
		t.Errorf("kafkaMessageToEntity should not return error, got: %v", err)
	}

	if result.ID != "trans-id-123" {
		t.Errorf("Expected ID 'trans-id-123', got %s", result.ID)
	}
	if result.TransactionType != entities.TransactionTypePayment {
		t.Errorf("Expected type PAYMENT, got %s", result.TransactionType)
	}
	if result.TransactionStatus != entities.TransactionStatusSuccess {
		t.Errorf("Expected status SUCCESS, got %s", result.TransactionStatus)
	}
	if result.Description == nil || *result.Description != "Test payment" {
		t.Error("Description should be set")
	}
	if result.ExternalReference == nil || *result.ExternalReference != "ext-ref-123" {
		t.Error("External reference should be set")
	}
	if result.PaymentMethod == nil || *result.PaymentMethod != "BANK_TRANSFER" {
		t.Error("Payment method should be set")
	}
	if result.Metadata == nil || *result.Metadata != `{"key": "value"}` {
		t.Error("Metadata should be set")
	}
}

func TestTransactionHandler_kafkaMessageToEntity_EmptyOptionalFields(t *testing.T) {
	mockUseCase := &mockTransactionUseCase{}
	mockLog := &mockLogger{}
	handler := NewTransactionHandler(mockUseCase, mockLog)

	kafkaMsg := &KafkaTransactionMessage{
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
		Description:              "", // Empty description
		PaymentMethod:            "", // Empty payment method
		IsAccessibleFromExternal: false,
		CreatedAt:                []interface{}{2024.0, 1.0, 1.0, 12.0, 0.0, 0.0},
		UpdatedAt:                []interface{}{2024.0, 1.0, 1.0, 12.0, 0.0, 0.0},
	}

	result, err := handler.kafkaMessageToEntity(kafkaMsg)
	if err != nil {
		t.Errorf("kafkaMessageToEntity should not return error, got: %v", err)
	}

	// Check that optional fields are not set when empty
	if result.Description != nil {
		t.Error("Description should be nil for empty description")
	}
	if result.PaymentMethod != nil {
		t.Error("PaymentMethod should be nil for empty payment method")
	}
	if result.IsAccessibleFromExternal != false {
		t.Error("IsAccessibleFromExternal should be false")
	}
}

func TestTransactionHandler_kafkaMessageToEntity_InvalidTimestamp(t *testing.T) {
	mockUseCase := &mockTransactionUseCase{}
	mockLog := &mockLogger{}
	handler := NewTransactionHandler(mockUseCase, mockLog)

	kafkaMsg := &KafkaTransactionMessage{
		ID:                       "trans-id-123",
		UserID:                   456,
		AccountID:                "account-456",
		TransactionID:            "trans-456",
		TransactionType:          "TOPUP",
		TransactionStatus:        "SUCCESS",
		Amount:                   100.00,
		IsAccessibleFromExternal: true,
		CreatedAt:                []interface{}{2024.0}, // Invalid timestamp (too short)
		UpdatedAt:                []interface{}{2024.0, 1.0, 1.0, 12.0, 0.0, 0.0},
	}

	result, err := handler.kafkaMessageToEntity(kafkaMsg)
	if err != nil {
		t.Errorf("kafkaMessageToEntity should not return error even with invalid timestamp, got: %v", err)
	}

	// Should use current time when timestamp parsing fails
	if result.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero even with invalid timestamp")
	}
}

func TestKafkaTransactionMessage_AllFields(t *testing.T) {
	// Test that the struct has all expected fields
	msg := KafkaTransactionMessage{
		ID:                       "test-id",
		UserID:                   123,
		AccountID:                "account-123",
		TransactionID:            "trans-123",
		TransactionType:          "TOPUP",
		TransactionStatus:        "SUCCESS",
		Amount:                   100.0,
		BalanceBefore:            1000.0,
		BalanceAfter:             1100.0,
		Currency:                 "IDR",
		Description:              "test",
		PaymentMethod:            "GOPAY",
		IsAccessibleFromExternal: true,
		CreatedAt:                []interface{}{2024.0, 1.0, 1.0, 0.0, 0.0, 0.0},
		UpdatedAt:                []interface{}{2024.0, 1.0, 1.0, 0.0, 0.0, 0.0},
	}

	// Test JSON marshaling/unmarshaling
	data, err := json.Marshal(msg)
	if err != nil {
		t.Errorf("Failed to marshal KafkaTransactionMessage: %v", err)
	}

	var unmarshaled KafkaTransactionMessage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal KafkaTransactionMessage: %v", err)
	}

	if unmarshaled.ID != msg.ID {
		t.Errorf("Expected ID %s, got %s", msg.ID, unmarshaled.ID)
	}
	if unmarshaled.TransactionType != msg.TransactionType {
		t.Errorf("Expected TransactionType %s, got %s", msg.TransactionType, unmarshaled.TransactionType)
	}
}

func TestTransactionHandler_HandleMessage_AllTransactionTypes(t *testing.T) {
	transactionTypes := []string{"TOPUP", "PAYMENT", "REFUND", "TRANSFER"}

	for _, txType := range transactionTypes {
		t.Run(txType, func(t *testing.T) {
			mockUseCase := &mockTransactionUseCase{}
			mockLog := &mockLogger{}
			handler := NewTransactionHandler(mockUseCase, mockLog)

			kafkaMsg := KafkaTransactionMessage{
				ID:                       "trans-id-" + txType,
				UserID:                   456,
				AccountID:                "account-456",
				TransactionID:            "trans-456-" + txType,
				TransactionType:          txType,
				TransactionStatus:        "SUCCESS",
				Amount:                   250.75,
				BalanceBefore:            1000.00,
				BalanceAfter:             1250.75,
				Currency:                 "IDR",
				IsAccessibleFromExternal: true,
				CreatedAt:                []interface{}{2024.0, 1.0, 15.0, 10.0, 30.0, 45.0},
				UpdatedAt:                []interface{}{2024.0, 1.0, 15.0, 10.0, 30.0, 45.0},
			}

			message, _ := json.Marshal(kafkaMsg)

			ctx := context.Background()
			err := handler.HandleMessage(ctx, message)

			if err != nil {
				t.Errorf("HandleMessage should not return error for %s, got: %v", txType, err)
			}

			if len(mockUseCase.processed) != 1 {
				t.Errorf("Expected 1 processed transaction for %s, got %d", txType, len(mockUseCase.processed))
			}

			processedTx := mockUseCase.processed[0]
			if string(processedTx.TransactionType) != txType {
				t.Errorf("Expected transaction type %s, got %s", txType, processedTx.TransactionType)
			}
		})
	}
}

func TestTransactionHandler_HandleMessage_AllTransactionStatuses(t *testing.T) {
	statuses := []string{"PENDING", "SUCCESS", "FAILED", "CANCELLED"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			mockUseCase := &mockTransactionUseCase{}
			mockLog := &mockLogger{}
			handler := NewTransactionHandler(mockUseCase, mockLog)

			kafkaMsg := KafkaTransactionMessage{
				ID:                       "trans-id-" + status,
				UserID:                   456,
				AccountID:                "account-456",
				TransactionID:            "trans-456-" + status,
				TransactionType:          "TOPUP",
				TransactionStatus:        status,
				Amount:                   250.75,
				BalanceBefore:            1000.00,
				BalanceAfter:             1250.75,
				Currency:                 "IDR",
				IsAccessibleFromExternal: true,
				CreatedAt:                []interface{}{2024.0, 1.0, 15.0, 10.0, 30.0, 45.0},
				UpdatedAt:                []interface{}{2024.0, 1.0, 15.0, 10.0, 30.0, 45.0},
			}

			message, _ := json.Marshal(kafkaMsg)

			ctx := context.Background()
			err := handler.HandleMessage(ctx, message)

			if err != nil {
				t.Errorf("HandleMessage should not return error for %s, got: %v", status, err)
			}

			if len(mockUseCase.processed) != 1 {
				t.Errorf("Expected 1 processed transaction for %s, got %d", status, len(mockUseCase.processed))
			}

			processedTx := mockUseCase.processed[0]
			if string(processedTx.TransactionStatus) != status {
				t.Errorf("Expected transaction status %s, got %s", status, processedTx.TransactionStatus)
			}
		})
	}
}
