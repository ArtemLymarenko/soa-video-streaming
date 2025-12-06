package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"soa-video-streaming/pkg/saga"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewSagaRepository,
		),
	)
}

type SagaRepository struct {
	pool *pgxpool.Pool
}

func NewSagaRepository(pool *pgxpool.Pool) *SagaRepository {
	return &SagaRepository{pool: pool}
}

func (r *SagaRepository) Create(ctx context.Context, correlationID string, status saga.SagaStateStatus, data json.RawMessage) (*saga.SagaStateEntity, error) {
	sagaState := &saga.SagaStateEntity{
		ID:            uuid.NewString(),
		CorrelationID: correlationID,
		Status:        status,
		Data:          data,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	var dataJSON []byte
	if data != nil {
		dataJSON = data
	}

	query := `
		INSERT INTO orchestrator_service.saga_state (id, correlation_id, state, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	var err error
	_, err = r.pool.Exec(ctx, query,
		sagaState.ID,
		sagaState.CorrelationID,
		sagaState.Status,
		dataJSON,
		sagaState.CreatedAt,
		sagaState.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return sagaState, nil
}

func (r *SagaRepository) Update(ctx context.Context, correlationID string, status saga.SagaStateStatus, data json.RawMessage) error {
	var dataJSON []byte
	if data != nil {
		dataJSON = data
	}

	query := `
		UPDATE orchestrator_service.saga_state
		SET state = $1, data = $2, updated_at = $3
		WHERE correlation_id = $4
	`

	_, err := r.pool.Exec(ctx, query, status, dataJSON, time.Now(), correlationID)
	return err
}

func (r *SagaRepository) Complete(ctx context.Context, correlationID string) error {
	query := `
		UPDATE orchestrator_service.saga_state
		SET state = $1, completed_at = $2, updated_at = $3
		WHERE correlation_id = $4
	`

	_, err := r.pool.Exec(ctx, query, saga.SagaStateCompleted, time.Now(), time.Now(), correlationID)
	return err
}

func (r *SagaRepository) FindByCorrelationID(ctx context.Context, correlationID string) (*saga.SagaStateEntity, error) {
	query := `
		SELECT id, correlation_id, state, data, created_at, updated_at, completed_at
		FROM orchestrator_service.saga_state
		WHERE correlation_id = $1
	`

	var sagaState saga.SagaStateEntity
	var dataJSON []byte
	var completedAt *time.Time

	err := r.pool.QueryRow(ctx, query, correlationID).Scan(
		&sagaState.ID,
		&sagaState.CorrelationID,
		&sagaState.Status,
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

func (r *SagaRepository) GetSteps(ctx context.Context, sagaStateID string) ([]saga.SagaStep, error) {
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

	var steps []saga.SagaStep
	for rows.Next() {
		var step saga.SagaStep
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
