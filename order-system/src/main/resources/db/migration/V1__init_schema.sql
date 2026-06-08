CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    amount NUMERIC(19, 2) NOT NULL,
    status VARCHAR(20) NOT NULL,
    timeout_at TIMESTAMP NOT NULL,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    version BIGINT DEFAULT 0
);

CREATE INDEX idx_orders_status_timeout ON orders(status, timeout_at);
CREATE INDEX idx_orders_order_no ON orders(order_no);
