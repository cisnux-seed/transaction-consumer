package consumer

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"time"
	"transaction-consumer/internal/infrastructures/config"
	"transaction-consumer/pkg/logger"
)

// Consumer represents Kafka consumer
type Consumer struct {
	reader *kafka.Reader
	logger logger.Logger
}

// MessageHandler defines the function signature for message handling
type MessageHandler func(ctx context.Context, message []byte) error

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg config.KafkaConfig, log logger.Logger) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		Topic:          cfg.Topic,
		MaxBytes:       cfg.MaxBytes,
		CommitInterval: cfg.CommitInterval,
		StartOffset:    kafka.LastOffset,
		ErrorLogger:    kafka.LoggerFunc(log.Error),
	})

	return &Consumer{
		reader: reader,
		logger: log,
	}, nil
}

// Consume starts consuming messages
func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	c.logger.Info("Starting Kafka consumer", "topic", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Consumer context cancelled, stopping...")
			return ctx.Err()
		default:
			message, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				c.logger.Error("Failed to fetch message", "error", err)
				time.Sleep(time.Second) // Backoff
				continue
			}

			// Process message
			if err := handler(ctx, message.Value); err != nil {
				c.logger.Error("Failed to process message", "error", err)
				// Continue processing other messages
			}

			// Commit message
			if err := c.reader.CommitMessages(ctx, message); err != nil {
				c.logger.Error("Failed to commit message", "error", err)
			}
		}
	}
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}
