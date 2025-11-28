CREATE TABLE IF NOT EXISTS user_service.outbox (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    metadata BYTEA,
    payload BYTEA NOT NULL,
    times_attempted INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_outbox_created_at ON user_service.outbox (created_at);
CREATE INDEX IF NOT EXISTS idx_outbox_scheduled_at ON user_service.outbox (scheduled_at);
