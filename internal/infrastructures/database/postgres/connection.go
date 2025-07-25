package postgres

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
	"transaction-consumer/internal/infrastructures/config"
)

// NewConnection creates a new database connection
func NewConnection(cfg config.DatabaseConfig, appConfig config.AppConfig) (*gorm.DB, error) {
	// Use the config's DSN method
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode)

	// Configure GORM logger level based on app environment and log level
	var gormLogLevel logger.LogLevel
	if appConfig.Environment == "development" || appConfig.Debug {
		switch appConfig.LogLevel {
		case "debug":
			gormLogLevel = logger.Info
		case "info":
			gormLogLevel = logger.Warn
		default:
			gormLogLevel = logger.Error
		}
	} else {
		gormLogLevel = logger.Error // Production: only errors
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool with values from config
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// CloseConnection closes the database connection
func CloseConnection(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
