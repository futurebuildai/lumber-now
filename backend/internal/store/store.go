package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

type Store struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		Pool:    pool,
		Queries: db.New(pool),
	}
}

type PoolConfig struct {
	MaxConns int32
	MinConns int32
}

func NewPool(ctx context.Context, databaseURL string, opts ...PoolConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	maxConns := int32(25)
	minConns := int32(5)
	if len(opts) > 0 {
		if opts[0].MaxConns > 0 {
			maxConns = opts[0].MaxConns
		}
		if opts[0].MinConns > 0 {
			minConns = opts[0].MinConns
		}
	}
	config.MaxConns = maxConns
	config.MinConns = minConns
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func (s *Store) WithTx(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)
	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
