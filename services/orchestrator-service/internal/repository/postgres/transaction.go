package postgres

import (
	"context"
	"soa-video-streaming/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type TransactionManager struct {
	client *postgres.Client
}

func NewTransactionManager(client *postgres.Client) *TransactionManager {
	return &TransactionManager{
		client: client,
	}
}

func (tm *TransactionManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return tm.client.Tx(ctx, func(tx pgx.Tx) error {
		return fn(ctx, tx)
	})
}
