-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(60) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    balance DOUBLE PRECISION DEFAULT 0
);
CREATE TABLE IF NOT EXISTS balance_withdrawals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    order_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP
);
CREATE TABLE IF NOT EXISTS user_orders (
    order_id VARCHAR(255) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    status VARCHAR(40) NOT NULL,
    accrual DOUBLE PRECISION,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS balance_withdrawals;
DROP TABLE IF EXISTS user_orders;
-- +goose StatementEnd
