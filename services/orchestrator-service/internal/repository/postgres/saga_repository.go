package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/orchestrator-service/internal/domain/entity"
)

type SagaRepository struct {
	pool *pgxpool.Pool
}

func NewSagaRepository(pool *pgxpool.Pool) *SagaRepository {
	return &SagaRepository{pool: pool}
}

// Create creates a new saga state
func (r *SagaRepository) Create(ctx context.Context, correlationID string, state saga.SagaState, data map[string]interface{}) (*entity.SagaState, error) {
	sagaState := &entity.SagaState{
		ID:            uuid.NewString(),
		CorrelationID: correlationID,
		State:         state,
		Data:          data,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO orchestrator_service.saga_state (id, correlation_id, state, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = r.pool.Exec(ctx, query,
		sagaState.ID,
		sagaState.CorrelationID,
		sagaState.State,
		dataJSON,
		sagaState.CreatedAt,
		sagaState.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return sagaState, nil
}

// Update updates the saga state
func (r *SagaRepository) Update(ctx context.Context, correlationID string, state saga.SagaState, data map[string]interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	query := `
		UPDATE orchestrator_service.saga_state
		SET state = $1, data = $2, updated_at = $3
		WHERE correlation_id = $4
	`

	_, err = r.pool.Exec(ctx, query, state, dataJSON, time.Now(), correlationID)
	return err
}

// Complete marks the saga as completed
func (r *SagaRepository) Complete(ctx context.Context, correlationID string) error {
	query := `
		UPDATE orchestrator_service.saga_state
		SET state = $1, completed_at = $2, updated_at = $3
		WHERE correlation_id = $4
	`

	_, err := r.pool.Exec(ctx, query, saga.SagaStateCompleted, time.Now(), time.Now(), correlationID)
	return err
}

// FindByCorrelationID finds a saga by correlation ID
func (r *SagaRepository) FindByCorrelationID(ctx context.Context, correlationID string) (*entity.SagaState, error) {
	query := `
		SELECT id, correlation_id, state, data, created_at, updated_at, completed_at
		FROM orchestrator_service.saga_state
		WHERE correlation_id = $1
	`

	var sagaState entity.SagaState
	var dataJSON []byte
	var completedAt *time.Time

	err := r.pool.QueryRow(ctx, query, correlationID).Scan(
		&sagaState.ID,
		&sagaState.CorrelationID,
		&sagaState.State,
		&dataJSON,
		&sagaState.CreatedAt,
		&sagaState.UpdatedAt,
		&completedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	sagaState.CompletedAt = completedAt

	if len(dataJSON) > 0 {
		err = json.Unmarshal(dataJSON, &sagaState.Data)
		if err != nil {
			return nil, err
		}
	}

	return &sagaState, nil
}

// AddStep adds a step to the saga
func (r *SagaRepository) AddStep(ctx context.Context, sagaStateID, stepName, serviceName string, status saga.StepStatus) error {
	query := `
		INSERT INTO orchestrator_service.saga_steps (id, saga_state_id, step_name, service_name, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		uuid.NewString(),
		sagaStateID,
		stepName,
		serviceName,
		status,
		time.Now(),
	)

	return err
}

// UpdateStep updates a saga step
func (r *SagaRepository) UpdateStep(ctx context.Context, sagaStateID, stepName string, status saga.StepStatus, errorMessage string) error {
	now := time.Now()

	query := `
		UPDATE orchestrator_service.saga_steps
		SET status = $1, error_message = $2, executed_at = $3
		WHERE saga_state_id = $4 AND step_name = $5
	`

	_, err := r.pool.Exec(ctx, query, status, errorMessage, now, sagaStateID, stepName)
	return err
}

// GetSteps retrieves all steps for a saga
func (r *SagaRepository) GetSteps(ctx context.Context, sagaStateID string) ([]entity.SagaStep, error) {
	query := `
		SELECT id, saga_state_id, step_name, service_name, status, executed_at, compensated_at, error_message, created_at
		FROM orchestrator_service.saga_steps
		WHERE saga_state_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, sagaStateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []entity.SagaStep
	for rows.Next() {
		var step entity.SagaStep
		err := rows.Scan(
			&step.ID,
			&step.SagaStateID,
			&step.StepName,
			&step.ServiceName,
			&step.Status,
			&step.ExecutedAt,
			&step.CompensatedAt,
			&step.ErrorMessage,
			&step.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	return steps, rows.Err()
}
