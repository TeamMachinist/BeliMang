-- Enable pg_trgm extension for efficient text searching
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    product_category VARCHAR(10) NOT NULL,
    price BIGINT NOT NULL,  
    image_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NUll DEFAULT NOW() 
);

CREATE INDEX IF NOT EXISTS idx_items_merchant_id ON items (merchant_id);
CREATE INDEX IF NOT EXISTS idx_items_product_category ON items (product_category);
CREATE INDEX IF NOT EXISTS idx_items_created_at ON items (created_at);
CREATE INDEX IF NOT EXISTS idx_items_merchant_id_product_category ON items (merchant_id, product_category);
CREATE INDEX IF NOT EXISTS idx_items_merchant_id_created_at ON items (merchant_id, created_at);
-- For text search operations using pg_trgm:
CREATE INDEX IF NOT EXISTS idx_items_name_text_search ON items USING gin (name gin_trgm_ops);
