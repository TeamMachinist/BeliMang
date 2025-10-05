-- Enable pg_trgm extension if not already enabled
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    estimate_id UUID NOT NULL REFERENCES estimates(id) ON DELETE CASCADE,
    total_price BIGINT NOT NULL,
    estimated_delivery_time_in_minutes INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create order_merchants table to store the relationship between orders and merchants
CREATE TABLE order_merchants (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    is_starting_point BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create order_items table to store the items for each order
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    order_merchant_id UUID NOT NULL REFERENCES order_merchants(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for filtering orders by user_id (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

-- Index for orders with user_id + created_at DESC (pagination pattern)
CREATE INDEX IF NOT EXISTS idx_orders_user_created_desc ON orders(user_id, created_at DESC NULLS LAST);

-- Index for order_merchants lookups
CREATE INDEX IF NOT EXISTS idx_order_merchants_order_id ON order_merchants(order_id);
CREATE INDEX IF NOT EXISTS idx_order_merchants_merchant_id ON order_merchants(merchant_id);

-- Composite index for order_merchants with order_id + created_at (maintain merchant order)
CREATE INDEX IF NOT EXISTS idx_order_merchants_order_created ON order_merchants(order_id, created_at ASC NULLS LAST);

-- Index for order_items lookups
CREATE INDEX IF NOT EXISTS idx_order_items_order_merchant_id ON order_items(order_merchant_id);
CREATE INDEX IF NOT EXISTS idx_order_items_item_id ON order_items(item_id);

-- Composite index for order_items with order_merchant_id + created_at (maintain item order)
CREATE INDEX IF NOT EXISTS idx_order_items_merchant_created ON order_items(order_merchant_id, created_at ASC NULLS LAST);