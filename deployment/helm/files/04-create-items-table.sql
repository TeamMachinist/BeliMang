\c belimang;
CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    product_category VARCHAR(10) NOT NULL,
    price BIGINT NOT NULL,
    image_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Create indexes
CREATE INDEX IF NOT EXISTS idx_items_merchant_id ON items(merchant_id);
CREATE INDEX IF NOT EXISTS idx_items_merchant_created_desc ON items(merchant_id, created_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_items_merchant_created_asc ON items(merchant_id, created_at ASC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_items_merchant_category_created_desc ON items(merchant_id, product_category, created_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_items_merchant_category_created_asc ON items(merchant_id, product_category, created_at ASC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_items_merchant_name_trgm ON items USING GIN(merchant_id, name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_items_id_merchant_covering ON items(id, merchant_id) INCLUDE (name, product_category, price, image_url, created_at);