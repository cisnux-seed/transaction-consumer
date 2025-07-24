package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"log"
	"strings"
	"time"
)

type Config struct {
	Kafka    KafkaConfig    `envPrefix:"KAFKA_"`
	Database DatabaseConfig `envPrefix:"DB_"`
	App      AppConfig      `envPrefix:"APP_"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers        []string      `env:"BROKERS,required" envSeparator:","`
	Topic          string        `env:"TOPIC,required"`
	GroupID        string        `env:"GROUP_ID,required"`
	CommitInterval time.Duration `env:"COMMIT_INTERVAL" envDefault:"2s"`
	MaxBytes       int           `env:"MAX_BYTES" envDefault:"10485760"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `env:"HOST,required"`
	Port            int           `env:"PORT" envDefault:"5432"`
	User            string        `env:"USER,required"`
	Password        string        `env:"PASSWORD,required"`
	Name            string        `env:"NAME,required"`
	SSLMode         string        `env:"SSLMODE" envDefault:"require"`
	MaxIdleConns    int           `env:"MAX_IDLE_CONNS" envDefault:"10"`
	MaxOpenConns    int           `env:"MAX_OPEN_CONNS" envDefault:"100"`
	ConnMaxLifetime time.Duration `env:"CONN_MAX_LIFETIME" envDefault:"1h"`
}

// AppConfig holds application configuration
type AppConfig struct {
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	Environment string `env:"ENVIRONMENT" envDefault:"production"`
	Port        int    `env:"PORT" envDefault:"8080"`
	Debug       bool   `env:"DEBUG" envDefault:"false"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Parse environment variables into the struct
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Additional validation
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Log configuration (without sensitive data)
	cfg.LogConfig()

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Kafka validation
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("KAFKA_BROKERS cannot be empty")
	}

	for i, broker := range c.Kafka.Brokers {
		c.Kafka.Brokers[i] = strings.TrimSpace(broker)
		if c.Kafka.Brokers[i] == "" {
			return fmt.Errorf("KAFKA_BROKERS contains empty broker at index %d", i)
		}
	}

	// Database validation
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("DB_PORT must be between 1 and 65535, got: %d", c.Database.Port)
	}

	validSSLModes := []string{"disable", "allow", "prefer", "require", "verify-ca", "verify-full"}
	if !contains(validSSLModes, c.Database.SSLMode) {
		return fmt.Errorf("DB_SSLMODE must be one of: %s, got: %s",
			strings.Join(validSSLModes, ", "), c.Database.SSLMode)
	}

	validLogLevels := []string{"debug", "info", "warn", "error", "fatal"}
	if !contains(validLogLevels, strings.ToLower(c.App.LogLevel)) {
		return fmt.Errorf("APP_LOG_LEVEL must be one of: %s, got: %s",
			strings.Join(validLogLevels, ", "), c.App.LogLevel)
	}

	return nil
}

// LogConfig logs the current configuration (without sensitive data)
func (c *Config) LogConfig() {
	log.Printf("Configuration loaded:")
	log.Printf("  Environment: %s", c.App.Environment)
	log.Printf("  Log Level: %s", c.App.LogLevel)
	log.Printf("  Port: %d", c.App.Port)
	log.Printf("  Debug: %t", c.App.Debug)
	log.Printf("  Kafka Brokers: %s", strings.Join(c.Kafka.Brokers, ", "))
	log.Printf("  Kafka Topic: %s", c.Kafka.Topic)
	log.Printf("  Kafka Group ID: %s", c.Kafka.GroupID)
	log.Printf("  Database Host: %s", c.Database.Host)
	log.Printf("  Database Port: %d", c.Database.Port)
	log.Printf("  Database Name: %s", c.Database.Name)
	log.Printf("  Database SSL Mode: %s", c.Database.SSLMode)
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.App.Environment) == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.App.Environment) == "production"
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		c.Database.Host, c.Database.User, c.Database.Password,
		c.Database.Name, c.Database.Port, c.Database.SSLMode)
}

// helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
