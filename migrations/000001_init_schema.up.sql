-- GoBilling Database Schema
-- All tables with relationships and constraints

-- ============================================================================
-- CORE DOMAIN TABLES
-- ============================================================================

-- Customers
CREATE TABLE IF NOT EXISTS customers (
    id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    metadata JSONB DEFAULT '{}',
    version INTEGER NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT customers_status_check CHECK (status IN ('active', 'suspended', 'deleted')),
    CONSTRAINT customers_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

CREATE INDEX idx_customers_email ON customers(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_status ON customers(status);
CREATE INDEX idx_customers_created_at ON customers(created_at);

-- Products
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_active ON products(active);
CREATE INDEX idx_products_name ON products(name);

-- Plans
CREATE TABLE IF NOT EXISTS plans (
    id VARCHAR(255) PRIMARY KEY,
    product_id VARCHAR(255) NOT NULL REFERENCES products(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    pricing_type VARCHAR(50) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    billing_interval VARCHAR(50) NOT NULL,
    billing_interval_count INTEGER NOT NULL DEFAULT 1,
    trial_period_days INTEGER NOT NULL DEFAULT 0,
    tiers JSONB,
    active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT plans_pricing_type_check CHECK (pricing_type IN ('flat', 'tiered', 'usage')),
    CONSTRAINT plans_billing_interval_check CHECK (billing_interval IN ('monthly', 'yearly')),
    CONSTRAINT plans_amount_check CHECK (amount >= 0),
    CONSTRAINT plans_interval_count_check CHECK (billing_interval_count > 0)
);

CREATE INDEX idx_plans_product_id ON plans(product_id);
CREATE INDEX idx_plans_active ON plans(active);

-- Subscriptions
CREATE TABLE IF NOT EXISTS subscriptions (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id),
    plan_id VARCHAR(255) NOT NULL REFERENCES plans(id),
    status VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    trial_start TIMESTAMP,
    trial_end TIMESTAMP,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    canceled_at TIMESTAMP,
    ended_at TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT subscriptions_status_check CHECK (status IN ('trialing', 'active', 'past_due', 'paused', 'canceled', 'expired')),
    CONSTRAINT subscriptions_quantity_check CHECK (quantity > 0)
);

CREATE INDEX idx_subscriptions_customer_id ON subscriptions(customer_id);
CREATE INDEX idx_subscriptions_plan_id ON subscriptions(plan_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_period_end ON subscriptions(current_period_end);

-- Invoices
CREATE SEQUENCE IF NOT EXISTS invoice_number_seq START 1000;

CREATE OR REPLACE FUNCTION generate_invoice_number()
RETURNS TEXT AS $$
DECLARE
    next_num BIGINT;
BEGIN
    next_num := nextval('invoice_number_seq');
    RETURN 'INV-' || LPAD(next_num::TEXT, 8, '0');
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS invoices (
    id VARCHAR(255) PRIMARY KEY,
    invoice_number VARCHAR(255) NOT NULL UNIQUE DEFAULT generate_invoice_number(),
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id),
    subscription_id VARCHAR(255) REFERENCES subscriptions(id),
    status VARCHAR(50) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    subtotal BIGINT NOT NULL DEFAULT 0,
    discount_amount BIGINT NOT NULL DEFAULT 0,
    tax_amount BIGINT NOT NULL DEFAULT 0,
    total BIGINT NOT NULL DEFAULT 0,
    amount_paid BIGINT NOT NULL DEFAULT 0,
    amount_due BIGINT NOT NULL DEFAULT 0,
    period_start TIMESTAMP,
    period_end TIMESTAMP,
    due_date TIMESTAMP,
    paid_at TIMESTAMP,
    voided_at TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT invoices_status_check CHECK (status IN ('draft', 'open', 'paid', 'void', 'uncollectible'))
);

CREATE INDEX idx_invoices_customer_id ON invoices(customer_id);
CREATE INDEX idx_invoices_subscription_id ON invoices(subscription_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);

-- Invoice Line Items
CREATE TABLE IF NOT EXISTS invoice_line_items (
    id VARCHAR(255) PRIMARY KEY,
    invoice_id VARCHAR(255) NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    quantity BIGINT NOT NULL,
    unit_amount BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    period_start TIMESTAMP,
    period_end TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoice_line_items_invoice_id ON invoice_line_items(invoice_id);

-- Payments
CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(255) PRIMARY KEY,
    invoice_id VARCHAR(255) NOT NULL REFERENCES invoices(id),
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL,
    payment_method_id VARCHAR(255),
    provider_id VARCHAR(255),
    failure_code VARCHAR(255),
    failure_message TEXT,
    idempotency_key VARCHAR(255) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT payments_status_check CHECK (status IN ('pending', 'processing', 'succeeded', 'failed', 'refunded'))
);

CREATE INDEX idx_payments_invoice_id ON payments(invoice_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE UNIQUE INDEX idx_payments_idempotency ON payments(invoice_id, idempotency_key);

-- Refunds
CREATE TABLE IF NOT EXISTS refunds (
    id VARCHAR(255) PRIMARY KEY,
    payment_id VARCHAR(255) NOT NULL REFERENCES payments(id),
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL,
    reason TEXT,
    provider_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT refunds_status_check CHECK (status IN ('pending', 'succeeded', 'failed'))
);

CREATE INDEX idx_refunds_payment_id ON refunds(payment_id);
CREATE INDEX idx_refunds_status ON refunds(status);

-- ============================================================================
-- FINANCIAL TRACKING
-- ============================================================================

-- Ledger Transactions
CREATE TABLE IF NOT EXISTS ledger_transactions (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id),
    type VARCHAR(50) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    invoice_id VARCHAR(255) REFERENCES invoices(id),
    payment_id VARCHAR(255) REFERENCES payments(id),
    refund_id VARCHAR(255) REFERENCES refunds(id),
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT ledger_type_check CHECK (type IN ('charge', 'payment', 'refund', 'credit', 'adjustment'))
);

CREATE INDEX idx_ledger_customer_id ON ledger_transactions(customer_id);
CREATE INDEX idx_ledger_type ON ledger_transactions(type);
CREATE INDEX idx_ledger_created_at ON ledger_transactions(created_at);

-- ============================================================================
-- EVENT & WEBHOOK SYSTEM
-- ============================================================================

-- Events (Transactional Outbox)
CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    retry_count INTEGER NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMP,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT events_status_check CHECK (status IN ('pending', 'processing', 'published', 'dead_letter'))
);

CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_next_retry ON events(next_retry_at) WHERE status = 'pending';

-- Webhook Endpoints
CREATE TABLE IF NOT EXISTS webhook_endpoints (
    id VARCHAR(255) PRIMARY KEY,
    url TEXT NOT NULL,
    secret VARCHAR(255) NOT NULL,
    events JSONB NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT webhook_url_check CHECK (url ~* '^https?://')
);

CREATE INDEX idx_webhook_endpoints_active ON webhook_endpoints(active);

-- Webhook Deliveries
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id VARCHAR(255) PRIMARY KEY,
    webhook_endpoint_id VARCHAR(255) NOT NULL REFERENCES webhook_endpoints(id),
    event_id VARCHAR(255) NOT NULL REFERENCES events(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    response_code INTEGER,
    response_body TEXT,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMP,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT webhook_deliveries_status_check CHECK (status IN ('pending', 'delivered', 'failed', 'skipped'))
);

CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX idx_webhook_deliveries_next_attempt ON webhook_deliveries(next_attempt_at) WHERE status = 'pending';

-- ============================================================================
-- INFRASTRUCTURE TABLES
-- ============================================================================

-- Idempotency Keys
CREATE TABLE IF NOT EXISTS idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    request_hash VARCHAR(255) NOT NULL,
    response_body TEXT,
    response_code INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_idempotency_expires ON idempotency_keys(expires_at);

-- Payment Retries
CREATE TABLE IF NOT EXISTS payment_retries (
    id VARCHAR(255) PRIMARY KEY,
    invoice_id VARCHAR(255) NOT NULL REFERENCES invoices(id),
    attempt_number INTEGER NOT NULL,
    scheduled_at TIMESTAMP NOT NULL,
    attempted_at TIMESTAMP,
    status VARCHAR(50) NOT NULL,
    payment_id VARCHAR(255) REFERENCES payments(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT payment_retries_status_check CHECK (status IN ('scheduled', 'attempted', 'succeeded', 'failed'))
);

CREATE INDEX idx_payment_retries_scheduled ON payment_retries(scheduled_at) WHERE status = 'scheduled';
CREATE INDEX idx_payment_retries_invoice ON payment_retries(invoice_id);

-- API Keys
CREATE TABLE IF NOT EXISTS api_keys (
    id VARCHAR(255) PRIMARY KEY,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(50) NOT NULL,
    name VARCHAR(255),
    permissions JSONB DEFAULT '[]',
    active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_active ON api_keys(active);
