-- Seed data for Pay2Go development environment
-- This script creates test partners and sample transactions

-- Clean existing test data (optional, comment out if you want to keep existing data)
-- TRUNCATE TABLE refunds, transaction_events, transactions, partners CASCADE;

-- Insert test partners with pre-generated API keys
-- API Key for partner 1: test-key-partner-1
-- API Key for partner 2: test-key-partner-2

INSERT INTO partners (id, name, email, api_key_hash, api_key_prefix, website, status, rate_limit_per_minute, created_at, updated_at)
VALUES 
    (
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid,
        'Acme Corporation',
        'payments@acme.com',
        '$2a$10$qVZ3vxDKJ1HYq5zQqG5c8.UhW0YqJZ9VqS7K8R4xN6fY7M8L9P0Qa',  -- bcrypt hash of 'test-key-partner-1'
        'test-key',
        'https://acme.com',
        'active',
        100,
        NOW(),
        NOW()
    ),
    (
        'b1ffcd99-9c0b-4ef8-bb6d-6bb9bd380a22'::uuid,
        'TechStart Inc',
        'api@techstart.io',
        '$2a$10$vWA4wyELK2IZr6aRrH6d9.ViX1ZrKA0WrT8L9S5yO7gZ9N1M2Q3Rb',  -- bcrypt hash of 'test-key-partner-2'
        'test-key',
        'https://techstart.io',
        'active',
        100,
        NOW(),
        NOW()
    );

-- Insert sample transactions for testing
INSERT INTO transactions (
    id, partner_id, amount, currency, status, payment_method, payment_provider,
    description, idempotency_key, metadata, created_at, updated_at
)
VALUES
    -- Completed transaction
    (
        'c2aabbcc-1111-4ef8-bb6d-6bb9bd380a33'::uuid,
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid,
        10000,  -- $100.00
        'USD',
        'completed',
        'credit_card',
        'stripe',
        'Test order #1001',
        'seed-tx-1001',
        '{"order_id": "1001", "customer_email": "customer1@example.com"}'::jsonb,
        NOW() - INTERVAL '5 days',
        NOW() - INTERVAL '5 days'
    ),
    -- Pending transaction
    (
        'c3bbccdd-2222-4ef8-bb6d-6bb9bd380a44'::uuid,
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid,
        25000,  -- $250.00
        'USD',
        'pending',
        'credit_card',
        'stripe',
        'Test order #1002',
        'seed-tx-1002',
        '{"order_id": "1002", "customer_email": "customer2@example.com"}'::jsonb,
        NOW() - INTERVAL '2 hours',
        NOW() - INTERVAL '2 hours'
    ),
    -- Failed transaction
    (
        'c4ccddee-3333-4ef8-bb6d-6bb9bd380a55'::uuid,
        'b1ffcd99-9c0b-4ef8-bb6d-6bb9bd380a22'::uuid,
        5000,  -- $50.00
        'EUR',
        'failed',
        'debit_card',
        'paypal',
        'Test order #2001',
        'seed-tx-2001',
        '{"order_id": "2001", "customer_email": "customer3@example.com"}'::jsonb,
        NOW() - INTERVAL '1 day',
        NOW() - INTERVAL '1 day'
    ),
    -- Completed transaction (eligible for refund)
    (
        'c5ddeeff-4444-4ef8-bb6d-6bb9bd380a66'::uuid,
        'b1ffcd99-9c0b-4ef8-bb6d-6bb9bd380a22'::uuid,
        15000,  -- $150.00
        'USD',
        'completed',
        'digital_wallet',
        'stripe',
        'Test order #2002',
        'seed-tx-2002',
        '{"order_id": "2002", "customer_email": "customer4@example.com", "wallet_type": "apple_pay"}'::jsonb,
        NOW() - INTERVAL '10 days',
        NOW() - INTERVAL '10 days'
    );

-- Update completed transactions with completion timestamps and provider IDs
UPDATE transactions
SET 
    completed_at = updated_at,
    provider_transaction_id = CONCAT('mock_', payment_provider, '_', substring(id::text, 1, 8))
WHERE status = 'completed';

-- Update failed transaction with failure reason
UPDATE transactions
SET failure_reason = 'Insufficient funds'
WHERE status = 'failed';

-- Insert transaction events for completed transactions
INSERT INTO transaction_events (id, transaction_id, event_type, status, provider_response, created_at)
SELECT
    gen_random_uuid(),
    id,
    'payment_completed',
    'completed',
    '{"status": "success", "message": "Payment processed successfully"}'::jsonb,
    completed_at
FROM transactions
WHERE status = 'completed';

-- Insert a sample refund
INSERT INTO refunds (
    id, transaction_id, amount, currency, status, reason,
    provider_refund_id, created_at, updated_at
)
VALUES (
    'd6eeffaa-5555-4ef8-bb6d-6bb9bd380a77'::uuid,
    'c2aabbcc-1111-4ef8-bb6d-6bb9bd380a33'::uuid,  -- First completed transaction
    3000,  -- $30.00 partial refund
    'USD',
    'completed',
    'Customer requested partial refund',
    'mock_refund_stripe_c2aabbcc',
    NOW() - INTERVAL '3 days',
    NOW() - INTERVAL '3 days'
);

-- Update the refunded transaction
UPDATE transactions
SET 
    status = 'refunded',
    refunded_amount = 3000,
    updated_at = NOW() - INTERVAL '3 days'
WHERE id = 'c2aabbcc-1111-4ef8-bb6d-6bb9bd380a33'::uuid;

-- Insert audit logs for transparency
INSERT INTO audit_logs (id, entity_type, entity_id, action, actor_id, actor_type, changes, created_at)
VALUES
    (
        gen_random_uuid(),
        'transaction',
        'c2aabbcc-1111-4ef8-bb6d-6bb9bd380a33'::uuid,
        'created',
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid,
        'partner',
        '{"status": "pending"}'::jsonb,
        NOW() - INTERVAL '5 days'
    ),
    (
        gen_random_uuid(),
        'transaction',
        'c2aabbcc-1111-4ef8-bb6d-6bb9bd380a33'::uuid,
        'updated',
        'system',
        'system',
        '{"status": "completed", "provider_transaction_id": "mock_stripe_c2aabbcc"}'::jsonb,
        NOW() - INTERVAL '5 days'
    ),
    (
        gen_random_uuid(),
        'refund',
        'd6eeffaa-5555-4ef8-bb6d-6bb9bd380a77'::uuid,
        'created',
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid,
        'partner',
        '{"amount": 3000, "reason": "Customer requested partial refund"}'::jsonb,
        NOW() - INTERVAL '3 days'
    );

-- Verify seed data
SELECT 'Partners created:' as info, COUNT(*) as count FROM partners;
SELECT 'Transactions created:' as info, COUNT(*) as count FROM transactions;
SELECT 'Refunds created:' as info, COUNT(*) as count FROM refunds;
SELECT 'Audit logs created:' as info, COUNT(*) as count FROM audit_logs;

-- Display partner API keys for testing
SELECT 
    name,
    email,
    'Use these API keys for testing (unhashed):' as note,
    CASE 
        WHEN email = 'payments@acme.com' THEN 'test-key-partner-1'
        WHEN email = 'api@techstart.io' THEN 'test-key-partner-2'
    END as api_key
FROM partners
ORDER BY created_at;
