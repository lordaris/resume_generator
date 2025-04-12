package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// NewPostgres creates a new PostgreSQL connection pool with proper configuration
func NewPostgres(dbURL string) (*sqlx.DB, error) {
	// Open database connection
	db, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify the connection is working
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Ping the database to check connectivity
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	log.Info().Msg("Successfully connected to PostgreSQL database")
	return db, nil
}
