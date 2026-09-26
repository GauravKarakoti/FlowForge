package store

import (
	"context"
)

// Store defines the interface for persisting workflow and job state.
type Store interface {
	SaveJob(ctx context.Context, job interface{}) error
	GetJob(ctx context.Context, jobID string) (interface{}, error)
	// TODO: Add methods for updating job status, querying logs, and managing workers
}

type PostgresStore struct {
	// TODO: Add pgxpool or sql.DB dependency
}

func NewPostgresStore(connectionString string) (*PostgresStore, error) {
	// TODO: Initialize and return connection pool
	return &PostgresStore{}, nil
}

func (p *PostgresStore) SaveJob(ctx context.Context, job interface{}) error {
	// TODO: Write SQL INSERT statement
	return nil
}

func (p *PostgresStore) GetJob(ctx context.Context, jobID string) (interface{}, error) {
	// TODO: Write SQL SELECT statement
	return nil, nil
}