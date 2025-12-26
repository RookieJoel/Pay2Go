-- Migration: Initial Schema for Payment Orchestration System
-- Version: 000001
-- Description: Creates core tables for partners, transactions, refunds, and audit logs

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create ENUM types
CREATE TYPE transaction_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed',
    'cancelled',
    'refunded',
    'partially_refunded'
);

CREATE TYPE payment_provider AS ENUM (
    'stripe',
    'paypal',
    'adyen',
    'manual'
);

CREATE TYPE refund_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed'
);

CREATE TYPE event_type AS ENUM (
    'created',
    'payment_initiated',
    'payment_processing',
    'payment_completed',
    'payment_failed',
    'refund_initiated',
    'refund_completed',
    'webhook_received',
    'status_changed'
);

-- ============================================================================
-- PARTNERS TABLE
-- ============================================================================
CREATE TABLE partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    api_key_hash VARCHAR(255) NOT NULL UNIQUE,
    api_key_prefix VARCHAR(10) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    rate_limit_per_minute INTEGER NOT NULL DEFAULT 100,
    webhook_url VARCHAR(512),
    webhook_secret VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_partners_api_key_prefix ON partners(api_key_prefix) WHERE deleted_at IS NULL;
CREATE INDEX idx_partners_is_active ON partners(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_partners_email ON partners(email) WHERE deleted_at IS NULL;

-- ============================================================================
-- TRANSACTIONS TABLE
-- ============================================================================
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES partners(id),
    
    idempotency_key VARCHAR(255) NOT NULL,
    
    amount DECIMAL(19, 4) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    
    payment_method VARCHAR(50) NOT NULL,
    provider payment_provider NOT NULL,
    provider_transaction_id VARCHAR(255),
    provider_customer_id VARCHAR(255),
    
    status transaction_status NOT NULL DEFAULT 'pending',
    
    customer_email VARCHAR(255) NOT NULL,
    customer_name VARCHAR(255),
    customer_phone VARCHAR(50),
    
    description TEXT,
    metadata JSONB,
    
    ip_address INET,
    user_agent TEXT,
    request_id UUID,
    
    error_code VARCHAR(50),
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT unique_partner_idempotency UNIQUE(partner_id, idempotency_key),
    CONSTRAINT check_amount_positive CHECK (amount >= 0.01),
    CONSTRAINT check_amount_max CHECK (amount <= 100000.00)
);

CREATE INDEX idx_transactions_partner_id ON transactions(partner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_status ON transactions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_provider_transaction_id ON transactions(provider_transaction_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_customer_email ON transactions(customer_email) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_idempotency ON transactions(partner_id, idempotency_key);
CREATE INDEX idx_transactions_partner_status_created ON transactions(partner_id, status, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_metadata ON transactions USING GIN (metadata);

-- ============================================================================
-- REFUNDS TABLE
-- ============================================================================
CREATE TABLE refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    
    amount DECIMAL(19, 4) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL,
    
    reason VARCHAR(255) NOT NULL,
    status refund_status NOT NULL DEFAULT 'pending',
    
    provider_refund_id VARCHAR(255),
    
    error_code VARCHAR(50),
    error_message TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT check_refund_amount_positive CHECK (amount > 0)
);

CREATE INDEX idx_refunds_transaction_id ON refunds(transaction_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_refunds_status ON refunds(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_refunds_created_at ON refunds(created_at DESC) WHERE deleted_at IS NULL;

-- ============================================================================
-- TRANSACTION EVENTS TABLE
-- ============================================================================
CREATE TABLE transaction_events (
    id BIGSERIAL PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    
    event_type event_type NOT NULL,
    status VARCHAR(50),
    
    provider_response JSONB,
    error_message TEXT,
    
    created_by VARCHAR(100),
    ip_address INET,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transaction_events_transaction_id ON transaction_events(transaction_id);
CREATE INDEX idx_transaction_events_created_at ON transaction_events(created_at DESC);
CREATE INDEX idx_transaction_events_type ON transaction_events(event_type);
CREATE INDEX idx_transaction_events_txn_created ON transaction_events(transaction_id, created_at DESC);

-- ============================================================================
-- AUDIT LOGS TABLE
-- ============================================================================
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    
    partner_id UUID REFERENCES partners(id),
    
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    
    ip_address INET NOT NULL,
    user_agent TEXT,
    request_id UUID,
    
    changes JSONB,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_partner_id ON audit_logs(partner_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);

-- ============================================================================
-- LOOKUP TABLES
-- ============================================================================
CREATE TABLE payment_methods (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

INSERT INTO payment_methods (code, name) VALUES
    ('card', 'Credit/Debit Card'),
    ('bank_transfer', 'Bank Transfer'),
    ('e_wallet', 'E-Wallet'),
    ('crypto', 'Cryptocurrency');

CREATE TABLE currencies (
    code VARCHAR(3) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(10),
    decimal_places SMALLINT NOT NULL DEFAULT 2,
    is_active BOOLEAN NOT NULL DEFAULT true
);

INSERT INTO currencies (code, name, symbol, decimal_places) VALUES
    ('USD', 'US Dollar', '$', 2),
    ('EUR', 'Euro', '€', 2),
    ('GBP', 'British Pound', '£', 2),
    ('JPY', 'Japanese Yen', '¥', 0),
    ('THB', 'Thai Baht', '฿', 2);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Updated At Trigger Function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_partners_updated_at 
    BEFORE UPDATE ON partners
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at 
    BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_refunds_updated_at 
    BEFORE UPDATE ON refunds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Transaction Event Logging Trigger
CREATE OR REPLACE FUNCTION log_transaction_changes()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO transaction_events (
        transaction_id,
        event_type,
        status,
        created_by
    ) VALUES (
        NEW.id,
        CASE 
            WHEN TG_OP = 'INSERT' THEN 'created'::event_type
            WHEN OLD.status != NEW.status THEN 'status_changed'::event_type
            ELSE 'status_changed'::event_type
        END,
        NEW.status::VARCHAR,
        'SYSTEM'
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER log_transaction_changes_trigger
    AFTER INSERT OR UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION log_transaction_changes();

-- Refund Amount Validation Trigger
CREATE OR REPLACE FUNCTION validate_refund_amount()
RETURNS TRIGGER AS $$
DECLARE
    total_refunded DECIMAL(19, 4);
    transaction_amount DECIMAL(19, 4);
BEGIN
    SELECT amount INTO transaction_amount
    FROM transactions
    WHERE id = NEW.transaction_id;
    
    SELECT COALESCE(SUM(amount), 0) INTO total_refunded
    FROM refunds
    WHERE transaction_id = NEW.transaction_id
    AND status = 'completed'
    AND id != NEW.id;
    
    IF (total_refunded + NEW.amount) > transaction_amount THEN
        RAISE EXCEPTION 'Refund amount exceeds transaction amount';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_refund_amount_trigger
    BEFORE INSERT OR UPDATE ON refunds
    FOR EACH ROW
    EXECUTE FUNCTION validate_refund_amount();

-- ============================================================================
-- COMMENTS (Documentation)
-- ============================================================================
COMMENT ON TABLE partners IS 'Stores partner/merchant information who use the payment orchestration API';
COMMENT ON TABLE transactions IS 'Core table storing all payment transactions';
COMMENT ON TABLE refunds IS 'Stores refund records linked to original transactions';
COMMENT ON TABLE transaction_events IS 'Audit trail for transaction lifecycle events';
COMMENT ON TABLE audit_logs IS 'Security and compliance audit logs for all operations';

COMMENT ON COLUMN transactions.idempotency_key IS 'Unique key per partner to prevent duplicate transactions';
COMMENT ON COLUMN transactions.metadata IS 'Flexible JSONB field for partner-specific custom data';
COMMENT ON COLUMN partners.api_key_hash IS 'Bcrypt hash of partner API key for authentication';
COMMENT ON COLUMN partners.api_key_prefix IS 'First 8 characters of API key for identification in logs';
