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
	poolConfig.MaxConnLifetime = time.Duration(cfg.ConnMaxLifetime) * time.Second
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.ConnectTimeout = 10 * time.Second
	poolConfig.AcquireTimeout = 30 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	log.Printf("[DB] Connected to PostgreSQL (pool: max=%d min=%d)", poolConfig.MaxConns, poolConfig.MinConns)
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

// Stats returns current pool statistics for monitoring
func Stats() *pgxpool.Stat {
	if pool == nil {
		return nil
	}
	stat := pool.Stat()
	return &stat
}
