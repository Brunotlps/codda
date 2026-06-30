CREATE TABLE orders (
    id          UUID PRIMARY KEY,
    status      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE order_items (
    order_id      UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    position      INTEGER NOT NULL,
    product_id    TEXT NOT NULL,
    product_name  TEXT NOT NULL,
    price_cents   BIGINT NOT NULL,
    quantity      INTEGER NOT NULL,
    PRIMARY KEY (order_id, position)
);

CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_status ON orders(status);
