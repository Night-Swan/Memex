package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect() error {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return fmt.Errorf("DATABASE_URL environment variable not set")
	}

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return fmt.Errorf("could not parse config: %w", err)
	}

	var poolErr error
	Pool, poolErr = pgxpool.NewWithConfig(context.Background(), config)
	if poolErr != nil {
		return fmt.Errorf("could not create pool: %w", poolErr)
	}

	if err := Pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("database unreachable: %w", err)
	}

	return nil
}