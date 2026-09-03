package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/trunglq04/goride/shared/env"

	_ "github.com/lib/pq"
)

// PostgresConfig holds PostgreSQL connection configuration.
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgresDefaultConfig creates a PostgresConfig from environment variables.
func NewPostgresDefaultConfig() *PostgresConfig {
	return &PostgresConfig{
		Host:     env.GetString("DB_HOST", "postgresql"),
		Port:     env.GetString("DB_PORT", "5432"),
		User:     env.GetString("DB_USER", "postgres"),
		Password: env.GetString("DB_PASSWORD", ""),
		DBName:   env.GetString("DB_NAME", "goride"),
		SSLMode:  env.GetString("DB_SSLMODE", "disable"),
	}
}

// DSN returns the PostgreSQL connection string.
func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// NewPostgresClient opens a PostgreSQL connection pool with sensible defaults.
func NewPostgresClient(cfg *PostgresConfig) (*sql.DB, error) {
	if cfg.Password == "" {
		return nil, fmt.Errorf("postgres password is required")
	}

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	slog.Info("Successfully connected to PostgreSQL",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.DBName,
	)

	return db, nil
}
