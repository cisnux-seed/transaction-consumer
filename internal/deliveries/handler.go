package deliveries

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"transaction-consumer/internal/domain/entities"
	"transaction-consumer/internal/usecases"
	"transaction-consumer/pkg/logger"
)

// TransactionHandler handles transaction messages from Kafka
type TransactionHandler struct {
	transactionUseCase usecases.TransactionUseCase
	logger             logger.Logger
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(uc usecases.TransactionUseCase, log logger.Logger) *TransactionHandler {
	return &TransactionHandler{
		transactionUseCase: uc,
		logger:             log,
	}
}

// KafkaTransactionMessage represents the incoming Kafka message structure
type KafkaTransactionMessage struct {
	ID                       string        `json:"id"`
	UserID                   int64         `json:"userId"`
	AccountID                string        `json:"accountId"`
	TransactionID            string        `json:"transactionId"`
	TransactionType          string        `json:"transactionType"`
	TransactionStatus        string        `json:"transactionStatus"`
	Amount                   float64       `json:"amount"`
	BalanceBefore            float64       `json:"balanceBefore"`
	BalanceAfter             float64       `json:"balanceAfter"`
	Currency                 string        `json:"currency"`
	Description              string        `json:"description"`
	ExternalReference        *string       `json:"externalReference"`
	PaymentMethod            string        `json:"paymentMethod"`
	Metadata                 *string       `json:"metadata"`
	IsAccessibleFromExternal bool          `json:"isAccessibleFromExternal"`
	CreatedAt                []interface{} `json:"createdAt"`
	UpdatedAt                []interface{} `json:"updatedAt"`
}

// HandleMessage handles incoming transaction messages
func (h *TransactionHandler) HandleMessage(ctx context.Context, message []byte) error {
	h.logger.Debug("Received message", "message", string(message))

	// Parse message
	var kafkaMsg KafkaTransactionMessage
	if err := json.Unmarshal(message, &kafkaMsg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	h.logger.Debug("Unmarshalled message", "message", kafkaMsg)

	// Convert to domain entities
	transaction, err := h.kafkaMessageToEntity(&kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed to convert message to entities: %w", err)
	}

	// Process transaction through use case
	if err := h.transactionUseCase.ProcessTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("failed to process transaction: %w", err)
	}

	return nil
}

// kafkaMessageToEntity converts Kafka message to domain entities
func (h *TransactionHandler) kafkaMessageToEntity(msg *KafkaTransactionMessage) (*entities.Transaction, error) {
	// Parse timestamps
	createdAt, err := h.parseTimestamp(msg.CreatedAt)
	if err != nil {
		h.logger.Warn("Failed to parse createdAt, using current time", "error", err)
		createdAt = time.Now().UTC()
	}

	updatedAt, err := h.parseTimestamp(msg.UpdatedAt)
	if err != nil {
		h.logger.Warn("Failed to parse updatedAt, using current time", "error", err)
		updatedAt = time.Now().UTC()
	}

	// Convert amount from cents to decimal
	amount := float64(msg.Amount) / 100.0

	transaction := &entities.Transaction{
		ID:                       msg.ID,
		UserID:                   msg.UserID,
		AccountID:                msg.AccountID,
		TransactionID:            msg.TransactionID,
		TransactionType:          entities.TransactionType(msg.TransactionType),
		TransactionStatus:        entities.TransactionStatus(msg.TransactionStatus),
		Amount:                   amount,
		BalanceBefore:            msg.BalanceBefore,
		BalanceAfter:             msg.BalanceAfter,
		Currency:                 msg.Currency,
		ExternalReference:        msg.ExternalReference,
		Metadata:                 msg.Metadata,
		IsAccessibleFromExternal: msg.IsAccessibleFromExternal,
		CreatedAt:                createdAt,
		UpdatedAt:                updatedAt,
	}

	// Set description if not empty
	if msg.Description != "" {
		transaction.Description = &msg.Description
	}

	// Set payment method if not empty
	if msg.PaymentMethod != "" {
		paymentMethod := entities.PaymentMethod(msg.PaymentMethod)
		transaction.PaymentMethod = &paymentMethod
	}

	return transaction, nil
}

// parseTimestamp converts array timestamp to time.Time
func (h *TransactionHandler) parseTimestamp(timestampArray []interface{}) (time.Time, error) {
	if len(timestampArray) < 6 {
		return time.Time{}, fmt.Errorf("invalid timestamp array length: %d", len(timestampArray))
	}

	year := int(timestampArray[0].(float64))
	month := int(timestampArray[1].(float64))
	day := int(timestampArray[2].(float64))
	hour := int(timestampArray[3].(float64))
	minute := int(timestampArray[4].(float64))
	second := int(timestampArray[5].(float64))

	var nanosecond int
	if len(timestampArray) > 6 {
		nanosecond = int(timestampArray[6].(float64))
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, time.UTC), nil
}
