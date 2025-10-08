\c belimang;
CREATE TABLE IF NOT EXISTS estimates (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_lat FLOAT8 NOT NULL,
    user_lng FLOAT8 NOT NULL,
    total_price BIGINT NOT NULL,
    estimated_delivery_time_in_minutes INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS estimate_orders (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    estimate_id UUID NOT NULL REFERENCES estimates(id) ON DELETE CASCADE,
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    is_starting_point BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS estimate_order_items (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    estimate_order_id UUID NOT NULL REFERENCES estimate_orders(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);