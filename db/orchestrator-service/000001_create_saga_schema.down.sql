DROP TRIGGER IF EXISTS update_saga_state_updated_at ON orchestrator_service.saga_state;
DROP FUNCTION IF EXISTS orchestrator_service.update_updated_at_column();
DROP TABLE IF EXISTS orchestrator_service.saga_steps;
DROP TABLE IF EXISTS orchestrator_service.saga_state;
DROP SCHEMA IF EXISTS orchestrator_service CASCADE;
