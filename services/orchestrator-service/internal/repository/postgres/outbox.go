package postgres

import (
	"context"
	"soa-video-streaming/pkg/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/oagudo/outbox"
)

type OutboxRepository struct {
	db postgres.DB
}

func NewOutboxRepository(db postgres.DB) *OutboxRepository {
	return &OutboxRepository{
		db: db,
	}
}

func (r *OutboxRepository) WithTx(tx pgx.Tx) *OutboxRepository {
	return &OutboxRepository{
		db: tx,
	}
}

const insertOutboxQuery = `
INSERT INTO orchestrator_service.outbox (id, created_at, scheduled_at, metadata, payload, times_attempted)
VALUES ($1, $2, $3, $4, $5, $6)
`

func (r *OutboxRepository) Save(ctx context.Context, msg *outbox.Message) error {
	_, err := r.db.Exec(ctx, insertOutboxQuery,
		msg.ID,
		msg.CreatedAt,
		msg.ScheduledAt,
		msg.Metadata,
		msg.Payload,
		msg.TimesAttempted,
	)
	return err
}
