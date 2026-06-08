ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS timeout_at TIMESTAMP WITH TIME ZONE;

UPDATE orders
SET timeout_at = created_at + INTERVAL '30 minutes'
WHERE timeout_at IS NULL;

ALTER TABLE orders
    ALTER COLUMN timeout_at SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_orders_status_timeout_at
    ON orders (status, timeout_at);
