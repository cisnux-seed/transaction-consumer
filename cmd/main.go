package main

import (
	"context"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"syscall"
	"time"
	"transaction-consumer/internal/infrastructures/config"
	"transaction-consumer/internal/infrastructures/database/postgres"
	"transaction-consumer/internal/usecases"
	"transaction-consumer/pkg/logger"

	kafkahandler "transaction-consumer/internal/deliveries"
	kafkainfra "transaction-consumer/internal/infrastructures/kafka/consumer"
)

func main() {
	// Initialize logger
	log := logger.NewLogger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Initialize database
	db, err := postgres.NewConnection(cfg.Database, cfg.App)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer func(db *gorm.DB) {
		err := postgres.CloseConnection(db)
		if err != nil {
			log.Error("Failed to close database connection", "error", err)
		} else {
			log.Info("Database connection closed successfully")
		}
	}(db)

	// Initialize repository
	transactionRepo := postgres.NewTransactionRepository(db, log)

	// Initialize use case
	transactionUsecase := usecases.NewTransactionUseCase(transactionRepo, log)

	// Initialize Kafka consumer
	kafkaConsumer, err := kafkainfra.NewConsumer(cfg.Kafka, log)
	if err != nil {
		log.Fatal("Failed to create Kafka consumer", "error", err)
	}
	defer func(kafkaConsumer *kafkainfra.Consumer) {
		err := kafkaConsumer.Close()
		if err != nil {
			log.Error("Failed to close Kafka consumer", "error", err)
		} else {
			log.Info("Kafka consumer closed successfully")
		}
	}(kafkaConsumer)

	// Initialize Kafka handler
	kafkaHandler := kafkahandler.NewTransactionHandler(transactionUsecase, log)

	// Start consuming
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consumer in goroutine
	go func() {
		if err := kafkaConsumer.Consume(ctx, kafkaHandler.HandleMessage); err != nil {
			log.Error("Kafka consumer error", "error", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down...")
	cancel()
	time.Sleep(2 * time.Second) // Grace period
}
