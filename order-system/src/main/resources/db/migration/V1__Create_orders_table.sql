CREATE TABLE orders (
    id              BIGSERIAL       PRIMARY KEY,
    order_no        VARCHAR(64)     NOT NULL,
    user_id         VARCHAR(64)     NOT NULL,
    amount          DECIMAL(12, 2)  NOT NULL,
    status          VARCHAR(32)     NOT NULL DEFAULT 'PENDING',
    timeout_at      TIMESTAMP,
    retry_count     INTEGER         NOT NULL DEFAULT 0,
    version         BIGINT,
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cancel_reason   VARCHAR(256)
);

CREATE UNIQUE INDEX idx_orders_order_no ON orders (order_no);
CREATE INDEX idx_orders_status_timeout ON orders (status, timeout_at);
