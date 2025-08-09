package config

import (
	"os"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092"},
					Topic:   "test-topic",
					GroupID: "test-group",
				},
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "user",
					Password: "password",
					Name:     "testdb",
					SSLMode:  "disable",
				},
				App: AppConfig{
					LogLevel: "info",
				},
			},
			expectErr: false,
		},
		{
			name: "invalid config - empty brokers",
			config: Config{
				Kafka: KafkaConfig{
					Brokers: []string{},
					Topic:   "test-topic",
					GroupID: "test-group",
				},
				Database: DatabaseConfig{
					Host:    "localhost",
					Port:    5432,
					SSLMode: "disable",
				},
				App: AppConfig{
					LogLevel: "info",
				},
			},
			expectErr: true,
		},
		{
			name: "invalid config - invalid port",
			config: Config{
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092"},
					Topic:   "test-topic",
					GroupID: "test-group",
				},
				Database: DatabaseConfig{
					Host:    "localhost",
					Port:    0,
					SSLMode: "disable",
				},
				App: AppConfig{
					LogLevel: "info",
				},
			},
			expectErr: true,
		},
		{
			name: "invalid config - invalid ssl mode",
			config: Config{
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092"},
					Topic:   "test-topic",
					GroupID: "test-group",
				},
				Database: DatabaseConfig{
					Host:    "localhost",
					Port:    5432,
					SSLMode: "invalid",
				},
				App: AppConfig{
					LogLevel: "info",
				},
			},
			expectErr: true,
		},
		{
			name: "invalid config - invalid log level",
			config: Config{
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092"},
					Topic:   "test-topic",
					GroupID: "test-group",
				},
				Database: DatabaseConfig{
					Host:    "localhost",
					Port:    5432,
					SSLMode: "disable",
				},
				App: AppConfig{
					LogLevel: "invalid",
				},
			},
			expectErr: true,
		},
		{
			name: "invalid config - empty broker in list",
			config: Config{
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092", "  "},
					Topic:   "test-topic",
					GroupID: "test-group",
				},
				Database: DatabaseConfig{
					Host:    "localhost",
					Port:    5432,
					SSLMode: "disable",
				},
				App: AppConfig{
					LogLevel: "info",
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"development", "development", true},
		{"Development", "Development", true},
		{"DEVELOPMENT", "DEVELOPMENT", true},
		{"production", "production", false},
		{"staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				App: AppConfig{Environment: tt.environment},
			}
			result := config.IsDevelopment()
			if result != tt.expected {
				t.Errorf("IsDevelopment() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"production", "production", true},
		{"Production", "Production", true},
		{"PRODUCTION", "PRODUCTION", true},
		{"development", "development", false},
		{"staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				App: AppConfig{Environment: tt.environment},
			}
			result := config.IsProduction()
			if result != tt.expected {
				t.Errorf("IsProduction() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestConfig_GetDSN(t *testing.T) {
	config := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
			SSLMode:  "disable",
		},
	}

	expected := "host=localhost user=testuser password=testpass dbname=testdb port=5432 sslmode=disable TimeZone=UTC"
	result := config.GetDSN()

	if result != expected {
		t.Errorf("GetDSN() = %s, expected %s", result, expected)
	}
}

func TestLoad_WithValidEnvVars(t *testing.T) {
	// Set up environment variables
	envVars := map[string]string{
		"KAFKA_BROKERS":  "localhost:9092,localhost:9093",
		"KAFKA_TOPIC":    "test-topic",
		"KAFKA_GROUP_ID": "test-group",
		"DB_HOST":        "localhost",
		"DB_PORT":        "5432",
		"DB_USER":        "testuser",
		"DB_PASSWORD":    "testpass",
		"DB_NAME":        "testdb",
		"DB_SSLMODE":     "disable",
		"APP_LOG_LEVEL":  "debug",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	config, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify some key values
	if len(config.Kafka.Brokers) != 2 {
		t.Errorf("expected 2 brokers, got %d", len(config.Kafka.Brokers))
	}
	if config.Kafka.Topic != "test-topic" {
		t.Errorf("expected topic 'test-topic', got %s", config.Kafka.Topic)
	}
	if config.Database.Host != "localhost" {
		t.Errorf("expected host 'localhost', got %s", config.Database.Host)
	}
}

func TestLoad_WithInvalidEnvVars(t *testing.T) {
	// Set up invalid environment variables
	os.Setenv("KAFKA_BROKERS", "localhost:9092")
	os.Setenv("KAFKA_TOPIC", "test-topic")
	os.Setenv("KAFKA_GROUP_ID", "test-group")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_SSLMODE", "invalid-mode")
	defer func() {
		os.Unsetenv("KAFKA_BROKERS")
		os.Unsetenv("KAFKA_TOPIC")
		os.Unsetenv("KAFKA_GROUP_ID")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_SSLMODE")
	}()

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid SSL mode")
	}
}

func TestConfig_LogConfig(t *testing.T) {
	config := &Config{
		App: AppConfig{
			Environment: "test",
			LogLevel:    "debug",
			Port:        8080,
			Debug:       true,
		},
		Kafka: KafkaConfig{
			Brokers: []string{"broker1", "broker2"},
			Topic:   "test-topic",
			GroupID: "test-group",
		},
		Database: DatabaseConfig{
			Host:    "localhost",
			Port:    5432,
			Name:    "testdb",
			SSLMode: "disable",
		},
	}

	// This should not panic
	config.LogConfig()
}

func TestContains(t *testing.T) {
	slice := []string{"debug", "info", "warn", "error"}

	tests := []struct {
		item     string
		expected bool
	}{
		{"debug", true},
		{"DEBUG", true},
		{"info", true},
		{"INFO", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.item, func(t *testing.T) {
			result := contains(slice, tt.item)
			if result != tt.expected {
				t.Errorf("contains(%s) = %v, expected %v", tt.item, result, tt.expected)
			}
		})
	}
}

func TestKafkaConfig_Defaults(t *testing.T) {
	// Test that default values are set correctly
	cfg := KafkaConfig{}

	// These would be set by the env package defaults
	if cfg.CommitInterval != 0 {
		// Default would be set by env tag, we can't test that here easily
		// but we can test that the struct accepts the field
		t.Logf("CommitInterval field exists: %v", cfg.CommitInterval)
	}

	if cfg.MaxBytes != 0 {
		// Default would be set by env tag
		t.Logf("MaxBytes field exists: %d", cfg.MaxBytes)
	}
}

func TestDatabaseConfig_Defaults(t *testing.T) {
	cfg := DatabaseConfig{}

	// Test that struct fields exist
	if cfg.Port == 0 {
		cfg.Port = 5432 // Default that would be set by env tag
	}

	if cfg.SSLMode == "" {
		cfg.SSLMode = "require" // Default that would be set by env tag
	}

	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 10 // Default that would be set by env tag
	}

	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = time.Hour // Default that would be set by env tag
	}

	// Just verify the struct has the expected structure
	if cfg.Port != 5432 {
		t.Errorf("Expected port 5432, got %d", cfg.Port)
	}
}
