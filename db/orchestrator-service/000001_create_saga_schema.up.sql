CREATE SCHEMA IF NOT EXISTS orchestrator_service;

-- Create saga_state table
CREATE TABLE IF NOT EXISTS orchestrator_service.saga_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    correlation_id UUID NOT NULL UNIQUE,
    state VARCHAR(20) NOT NULL,
    data JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- Create index on correlation_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_saga_state_correlation_id ON orchestrator_service.saga_state(correlation_id);
CREATE INDEX IF NOT EXISTS idx_saga_state_state ON orchestrator_service.saga_state(state);

-- Create saga_steps table
CREATE TABLE IF NOT EXISTS orchestrator_service.saga_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_state_id UUID NOT NULL REFERENCES orchestrator_service.saga_state(id) ON DELETE CASCADE,
    step_name VARCHAR(100) NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    executed_at TIMESTAMP,
    compensated_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for saga_steps
CREATE INDEX IF NOT EXISTS idx_saga_steps_saga_state_id ON orchestrator_service.saga_steps(saga_state_id);
CREATE INDEX IF NOT EXISTS idx_saga_steps_status ON orchestrator_service.saga_steps(status);

-- Create trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION orchestrator_service.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for automatic updated_at updates
CREATE TRIGGER update_saga_state_updated_at
    BEFORE UPDATE ON orchestrator_service.saga_state
    FOR EACH ROW
    EXECUTE FUNCTION orchestrator_service.update_updated_at_column();
