# Database Design - Payment Orchestration System

## 1. ENTITY-RELATIONSHIP DIAGRAM

```
┌─────────────────┐
│    partners     │
├─────────────────┤
│ id (PK)         │───┐
│ name            │   │
│ api_key_hash    │   │
│ is_active       │   │
│ created_at      │   │
│ updated_at      │   │
└─────────────────┘   │
                      │
                      │ 1:N
                      │
┌─────────────────────▼───────┐       ┌──────────────────┐
│      transactions           │       │  payment_methods │
├─────────────────────────────┤       ├──────────────────┤
│ id (PK)                     │       │ id (PK)          │
│ partner_id (FK)             │◄──────│ code             │
│ idempotency_key (UNIQUE)    │  N:1  │ name             │
│ amount                      │       │ is_active        │
│ currency                    │       └──────────────────┘
│ payment_method_id (FK)      │
│ provider                    │       ┌──────────────────┐
│ provider_transaction_id     │       │   currencies     │
│ status                      │       ├──────────────────┤
│ customer_email              │       │ code (PK)        │
│ customer_name               │       │ name             │
│ description                 │       │ symbol           │
│ metadata (JSONB)            │       │ decimal_places   │
│ created_at                  │       └──────────────────┘
│ updated_at                  │
│ processed_at                │
│ failed_at                   │
└─────────────────┬───────────┘
                  │
                  │ 1:N
                  │
┌─────────────────▼───────────┐
│         refunds             │
├─────────────────────────────┤
│ id (PK)                     │
│ transaction_id (FK)         │
│ amount                      │
│ reason                      │
│ status                      │
│ provider_refund_id          │
│ created_at                  │
│ updated_at                  │
│ processed_at                │
└─────────────────────────────┘

┌─────────────────────────────┐
│    transaction_events       │
├─────────────────────────────┤
│ id (PK)                     │
│ transaction_id (FK)         │
│ event_type                  │
│ status                      │
│ provider_response (JSONB)   │
│ error_message               │
│ created_at                  │
└─────────────────────────────┘

┌─────────────────────────────┐
│       audit_logs            │
├─────────────────────────────┤
│ id (PK)                     │
│ partner_id (FK)             │
│ action                      │
│ resource_type               │
│ resource_id                 │
│ ip_address                  │
│ user_agent                  │
│ request_id                  │
│ changes (JSONB)             │
│ created_at                  │
└─────────────────────────────┘
```

## 2. TABLE DEFINITIONS (PostgreSQL)

### 2.1 Partners Table
```sql
CREATE TABLE partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    api_key_hash VARCHAR(255) NOT NULL UNIQUE,
    api_key_prefix VARCHAR(10) NOT NULL, -- For identification (first 8 chars)
    is_active BOOLEAN NOT NULL DEFAULT true,
    rate_limit_per_minute INTEGER NOT NULL DEFAULT 100,
    webhook_url VARCHAR(512),
    webhook_secret VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance
CREATE INDEX idx_partners_api_key_prefix ON partners(api_key_prefix) WHERE deleted_at IS NULL;
CREATE INDEX idx_partners_is_active ON partners(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_partners_email ON partners(email) WHERE deleted_at IS NULL;
```

### 2.2 Transactions Table (Core Entity)
```sql
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

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES partners(id),
    
    -- Idempotency (prevent duplicate transactions)
    idempotency_key VARCHAR(255) NOT NULL,
    
    -- Financial details
    amount DECIMAL(19, 4) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    
    -- Payment details
    payment_method VARCHAR(50) NOT NULL, -- 'card', 'bank_transfer', 'e_wallet'
    provider payment_provider NOT NULL,
    provider_transaction_id VARCHAR(255),
    provider_customer_id VARCHAR(255),
    
    -- Transaction state
    status transaction_status NOT NULL DEFAULT 'pending',
    
    -- Customer information
    customer_email VARCHAR(255) NOT NULL,
    customer_name VARCHAR(255),
    customer_phone VARCHAR(50),
    
    -- Additional data
    description TEXT,
    metadata JSONB, -- Flexible field for partner-specific data
    
    -- Tracking
    ip_address INET,
    user_agent TEXT,
    request_id UUID,
    
    -- Error handling
    error_code VARCHAR(50),
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    
    -- Soft delete
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT unique_partner_idempotency UNIQUE(partner_id, idempotency_key),
    CONSTRAINT check_amount_positive CHECK (amount >= 0.01),
    CONSTRAINT check_amount_max CHECK (amount <= 100000.00)
);

-- Performance indexes
CREATE INDEX idx_transactions_partner_id ON transactions(partner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_status ON transactions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_provider_transaction_id ON transactions(provider_transaction_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_customer_email ON transactions(customer_email) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_idempotency ON transactions(partner_id, idempotency_key);

-- Composite index for common queries
CREATE INDEX idx_transactions_partner_status_created 
    ON transactions(partner_id, status, created_at DESC) 
    WHERE deleted_at IS NULL;

-- JSONB index for metadata queries
CREATE INDEX idx_transactions_metadata ON transactions USING GIN (metadata);

-- Partitioning preparation (for high-volume scenarios)
-- Could be partitioned by created_at (monthly partitions)
```

### 2.3 Refunds Table
```sql
CREATE TYPE refund_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed'
);

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

-- Indexes
CREATE INDEX idx_refunds_transaction_id ON refunds(transaction_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_refunds_status ON refunds(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_refunds_created_at ON refunds(created_at DESC) WHERE deleted_at IS NULL;
```

### 2.4 Transaction Events (Audit Trail)
```sql
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

CREATE TABLE transaction_events (
    id BIGSERIAL PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    
    event_type event_type NOT NULL,
    status VARCHAR(50),
    
    provider_response JSONB,
    error_message TEXT,
    
    created_by VARCHAR(100), -- System, partner, admin
    ip_address INET,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_transaction_events_transaction_id ON transaction_events(transaction_id);
CREATE INDEX idx_transaction_events_created_at ON transaction_events(created_at DESC);
CREATE INDEX idx_transaction_events_type ON transaction_events(event_type);

-- Composite index for event history queries
CREATE INDEX idx_transaction_events_txn_created 
    ON transaction_events(transaction_id, created_at DESC);
```

### 2.5 Audit Logs (Compliance & Security)
```sql
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    
    partner_id UUID REFERENCES partners(id),
    
    action VARCHAR(100) NOT NULL, -- 'create', 'update', 'delete', 'read'
    resource_type VARCHAR(50) NOT NULL, -- 'transaction', 'refund', 'partner'
    resource_id UUID,
    
    ip_address INET NOT NULL,
    user_agent TEXT,
    request_id UUID,
    
    changes JSONB, -- Before/after for updates
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for audit queries
CREATE INDEX idx_audit_logs_partner_id ON audit_logs(partner_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);

-- Time-series partitioning for audit logs (keeps table performant)
-- Partition by month
```

### 2.6 Payment Methods Lookup Table
```sql
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
```

### 2.7 Currencies Lookup Table
```sql
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
```

## 3. DATABASE FUNCTIONS & TRIGGERS

### 3.1 Updated At Trigger (DRY principle)
```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables with updated_at
CREATE TRIGGER update_partners_updated_at BEFORE UPDATE ON partners
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_refunds_updated_at BEFORE UPDATE ON refunds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

### 3.2 Audit Log Trigger (Auto-logging)
```sql
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
```

### 3.3 Refund Amount Validation
```sql
CREATE OR REPLACE FUNCTION validate_refund_amount()
RETURNS TRIGGER AS $$
DECLARE
    total_refunded DECIMAL(19, 4);
    transaction_amount DECIMAL(19, 4);
BEGIN
    -- Get original transaction amount
    SELECT amount INTO transaction_amount
    FROM transactions
    WHERE id = NEW.transaction_id;
    
    -- Calculate total refunded amount
    SELECT COALESCE(SUM(amount), 0) INTO total_refunded
    FROM refunds
    WHERE transaction_id = NEW.transaction_id
    AND status = 'completed'
    AND id != NEW.id;
    
    -- Check if refund exceeds transaction amount
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
```

## 4. PERFORMANCE OPTIMIZATION STRATEGIES

### 4.1 Connection Pooling
- Min connections: 10
- Max connections: 100
- Idle timeout: 30 seconds
- Max lifetime: 1 hour

### 4.2 Query Optimization
- Use EXPLAIN ANALYZE for slow queries
- Avoid N+1 queries (use JOINs or batch queries)
- Use prepared statements
- Limit result sets with pagination

### 4.3 Caching Strategy
- Cache partner details (Redis, TTL: 1 hour)
- Cache payment methods & currencies (TTL: 24 hours)
- Invalidate on update

### 4.4 Archival Strategy
- Archive transactions older than 2 years to separate table
- Keep transaction_events for 1 year
- Compress audit_logs after 90 days

## 5. DATA SECURITY

### 5.1 Encryption at Rest
```sql
-- Use pgcrypto extension for sensitive data
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Encrypt sensitive fields
-- Note: In production, use application-level encryption for PCI compliance
```

### 5.2 Row-Level Security (RLS)
```sql
-- Enable RLS for multi-tenant isolation
ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;

CREATE POLICY partner_isolation ON transactions
    FOR ALL
    USING (partner_id = current_setting('app.current_partner_id')::UUID);
```

### 5.3 Data Retention Policies
- Personal data (customer_email, customer_name): 7 years
- Transaction records: 10 years (compliance)
- Audit logs: 3 years

## 6. BACKUP & RECOVERY

- **Full backups**: Daily at 2 AM UTC
- **Incremental backups**: Every 6 hours
- **WAL archiving**: Enabled for point-in-time recovery
- **Retention**: 30 days
- **Testing**: Monthly restore tests

## 7. MONITORING QUERIES

### Slow Query Detection
```sql
-- Enable pg_stat_statements
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Find slow queries
SELECT query, calls, mean_exec_time, max_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

### Table Bloat Monitoring
```sql
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename))
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```
