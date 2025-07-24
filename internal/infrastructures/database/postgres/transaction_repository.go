package postgres

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"time"
	"transaction-consumer/internal/domain/entities"
	"transaction-consumer/internal/domain/repositories"
	"transaction-consumer/pkg/logger"
)

// TransactionModel represents the database model
type TransactionModel struct {
	ID                       string    `gorm:"primaryKey;type:varchar(36);default:gen_random_uuid()"`
	UserID                   int64     `gorm:"not null;index"`
	AccountID                string    `gorm:"not null;index;type:varchar(36)"`
	TransactionID            string    `gorm:"not null;uniqueIndex;type:varchar(50)"`
	TransactionType          string    `gorm:"not null;type:transaction_type_enum"`
	TransactionStatus        string    `gorm:"not null;index;type:transaction_status_enum"`
	Amount                   float64   `gorm:"not null;type:decimal(15,2)"`
	BalanceBefore            float64   `gorm:"not null;type:decimal(15,2)"`
	BalanceAfter             float64   `gorm:"not null;type:decimal(15,2)"`
	Currency                 string    `gorm:"not null;default:IDR;type:varchar(3)"`
	Description              *string   `gorm:"type:text"`
	ExternalReference        *string   `gorm:"type:varchar(255)"`
	PaymentMethod            *string   `gorm:"type:payment_method_enum"`
	Metadata                 *string   `gorm:"type:text"`
	IsAccessibleFromExternal bool      `gorm:"not null;default:true;column:is_accessible_external"`
	CreatedAt                time.Time `gorm:"not null;default:now()"`
	UpdatedAt                time.Time `gorm:"not null;default:now()"`
}

// TableName returns the table name
func (TransactionModel) TableName() string {
	return "historical_transactions"
}

// transactionRepository implements the repositories interface
type transactionRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewTransactionRepository creates a new transaction repositories
func NewTransactionRepository(db *gorm.DB, log logger.Logger) repositories.TransactionRepository {
	return &transactionRepository{
		db:     db,
		logger: log,
	}
}

// Create creates a new transaction
func (r *transactionRepository) Create(ctx context.Context, transaction *entities.Transaction) error {
	model := r.entityToModel(transaction)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update entities with generated ID
	transaction.ID = model.ID
	return nil
}

// GetByTransactionID retrieves a transaction by transaction ID
func (r *transactionRepository) GetByTransactionID(ctx context.Context, transactionID string) (*entities.Transaction, error) {
	var model TransactionModel

	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return r.modelToEntity(&model), nil
}

// Exists checks if a transaction exists by transaction ID
func (r *transactionRepository) Exists(ctx context.Context, transactionID string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&TransactionModel{}).Where("transaction_id = ?", transactionID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check transaction existence: %w", err)
	}

	return count > 0, nil
}

// entityToModel converts entities to database model
func (r *transactionRepository) entityToModel(transaction *entities.Transaction) *TransactionModel {
	model := &TransactionModel{
		ID:                       transaction.ID,
		UserID:                   transaction.UserID,
		AccountID:                transaction.AccountID,
		TransactionID:            transaction.TransactionID,
		TransactionType:          string(transaction.TransactionType),
		TransactionStatus:        string(transaction.TransactionStatus),
		Amount:                   transaction.Amount,
		BalanceBefore:            transaction.BalanceBefore,
		BalanceAfter:             transaction.BalanceAfter,
		Currency:                 transaction.Currency,
		Description:              transaction.Description,
		ExternalReference:        transaction.ExternalReference,
		Metadata:                 transaction.Metadata,
		IsAccessibleFromExternal: transaction.IsAccessibleFromExternal,
		CreatedAt:                transaction.CreatedAt,
		UpdatedAt:                transaction.UpdatedAt,
	}

	if transaction.PaymentMethod != nil {
		paymentMethod := string(*transaction.PaymentMethod)
		model.PaymentMethod = &paymentMethod
	}

	return model
}

// modelToEntity converts database model to entities
func (r *transactionRepository) modelToEntity(model *TransactionModel) *entities.Transaction {
	transaction := &entities.Transaction{
		ID:                       model.ID,
		UserID:                   model.UserID,
		AccountID:                model.AccountID,
		TransactionID:            model.TransactionID,
		TransactionType:          entities.TransactionType(model.TransactionType),
		TransactionStatus:        entities.TransactionStatus(model.TransactionStatus),
		Amount:                   model.Amount,
		BalanceBefore:            model.BalanceBefore,
		BalanceAfter:             model.BalanceAfter,
		Currency:                 model.Currency,
		Description:              model.Description,
		ExternalReference:        model.ExternalReference,
		Metadata:                 model.Metadata,
		IsAccessibleFromExternal: model.IsAccessibleFromExternal,
		CreatedAt:                model.CreatedAt,
		UpdatedAt:                model.UpdatedAt,
	}

	if model.PaymentMethod != nil {
		paymentMethod := entities.PaymentMethod(*model.PaymentMethod)
		transaction.PaymentMethod = &paymentMethod
	}

	return transaction
}
