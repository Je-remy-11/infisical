CREATE TABLE orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    status VARCHAR(20) NOT NULL,
    timeout_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    INDEX idx_status_timeout (status, timeout_at),
    INDEX idx_order_no (order_no),
    INDEX idx_user_id (user_id)
);
