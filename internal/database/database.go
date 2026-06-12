package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/CodingFervor/live-commerce-bi/internal/config"
)

var pool *pgxpool.Pool

func Init(cfg *config.DatabaseConfig) error {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return fmt.Errorf("parse db config: %w", err)
	}
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.HealthCheckPeriod = 30 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	log.Println("[DB] Connected to PostgreSQL")
	return nil
}

func Get() *pgxpool.Pool {
	return pool
}

func Close() {
	if pool != nil {
		pool.Close()
		log.Println("[DB] Connection pool closed")
	}
}
