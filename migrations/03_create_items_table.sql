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
