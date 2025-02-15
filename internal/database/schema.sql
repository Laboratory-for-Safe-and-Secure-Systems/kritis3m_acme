-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id VARCHAR(255) PRIMARY KEY,
    key JSONB NOT NULL,
    contact JSONB,
    status VARCHAR(50) NOT NULL,
    terms_agreed BOOLEAN NOT NULL DEFAULT false,
    created_at BIGINT NOT NULL,
    initial_ip VARCHAR(45) NOT NULL,
    created_at_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(255) PRIMARY KEY,
    account_id VARCHAR(255) NOT NULL REFERENCES accounts(id),
    status VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    not_before TIMESTAMP WITH TIME ZONE,
    not_after TIMESTAMP WITH TIME ZONE,
    identifiers JSONB NOT NULL,
    error JSONB,
    certificate_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Authorizations table
CREATE TABLE IF NOT EXISTS authorizations (
    id VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL REFERENCES orders(id),
    status VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    identifier JSONB NOT NULL,
    wildcard BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Challenges table
CREATE TABLE IF NOT EXISTS challenges (
    id VARCHAR(255) PRIMARY KEY,
    authorization_id VARCHAR(255) NOT NULL REFERENCES authorizations(id),
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    token VARCHAR(255) NOT NULL,
    validated_at TIMESTAMP WITH TIME ZONE,
    error JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Certificates table
CREATE TABLE IF NOT EXISTS certificates (
    id VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL REFERENCES orders(id),
    certificate TEXT NOT NULL,
    chain TEXT[],
    revoked BOOLEAN DEFAULT false,
    revocation_reason INT,
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes
CREATE INDEX idx_orders_account_id ON orders(account_id);
CREATE INDEX idx_authorizations_order_id ON authorizations(order_id);
CREATE INDEX idx_challenges_authorization_id ON challenges(authorization_id);
CREATE INDEX idx_certificates_order_id ON certificates(order_id); 