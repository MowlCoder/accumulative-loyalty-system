-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS good_rewards (
    id SERIAL PRIMARY KEY,
    match VARCHAR(255) NOT NULL UNIQUE,
    reward DOUBLE PRECISION NOT NULL,
    reward_type VARCHAR(10) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS registered_orders (
    order_id VARCHAR(255) PRIMARY KEY,
    status VARCHAR(255) NOT NULL,
    accrual DOUBLE PRECISION,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS orders_goods (
    order_id VARCHAR(255) REFERENCES registered_orders(order_id),
    description VARCHAR(512) NOT NULL,
    price DOUBLE PRECISION
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS good_rewards;
DROP TABLE IF EXISTS orders_goods;
DROP TABLE IF EXISTS registered_orders;
-- +goose StatementEnd
