package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"crawler/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	*pgxpool.Pool
}

func NewPool(ctx context.Context, cfg *config.Config) (*Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.GetConnectionURL())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	log.Println("Database pool created successfully")
	return &Pool{pool}, nil
}

func (p *Pool) Close() {
	if p.Pool != nil {
		p.Pool.Close()
		log.Println("Database pool closed")
	}
}
