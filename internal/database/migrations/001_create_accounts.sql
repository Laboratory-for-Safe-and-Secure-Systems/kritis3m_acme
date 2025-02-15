CREATE TABLE IF NOT EXISTS accounts (
    id VARCHAR(255) PRIMARY KEY,
    key JSONB NOT NULL,
    status VARCHAR(50) NOT NULL,
    contact TEXT[] DEFAULT '{}',
    terms_of_service_agreed BOOLEAN NOT NULL DEFAULT false,
    created_at BIGINT NOT NULL,
    initial_ip VARCHAR(255) NOT NULL,
    orders_url TEXT
); 