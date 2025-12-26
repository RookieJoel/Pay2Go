-- Rollback migration for initial schema

-- Drop triggers first
DROP TRIGGER IF EXISTS validate_refund_amount_trigger ON refunds;
DROP TRIGGER IF EXISTS log_transaction_changes_trigger ON transactions;
DROP TRIGGER IF EXISTS update_refunds_updated_at ON refunds;
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;
DROP TRIGGER IF EXISTS update_partners_updated_at ON partners;

-- Drop functions
DROP FUNCTION IF EXISTS validate_refund_amount();
DROP FUNCTION IF EXISTS log_transaction_changes();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order of creation due to foreign keys)
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS transaction_events CASCADE;
DROP TABLE IF EXISTS refunds CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS partners CASCADE;
DROP TABLE IF EXISTS currencies CASCADE;
DROP TABLE IF EXISTS payment_methods CASCADE;

-- Drop ENUM types
DROP TYPE IF EXISTS event_type;
DROP TYPE IF EXISTS refund_status;
DROP TYPE IF EXISTS payment_provider;
DROP TYPE IF EXISTS transaction_status;

-- Drop extensions (be careful in shared environments)
-- DROP EXTENSION IF EXISTS "pgcrypto";
-- DROP EXTENSION IF EXISTS "uuid-ossp";
