CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS btree_gin;  -- ADD THIS LINE

CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    product_category VARCHAR(10) NOT NULL,
    price BIGINT NOT NULL,
    image_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- 1. PRIMARY INDEX: merchant_id (base filter - always used)
CREATE INDEX IF NOT EXISTS idx_items_merchant_id 
ON items(merchant_id);

-- 2. COMPOSITE INDEXES for common query patterns
-- Pattern: merchant_id + created_at DESC (most common - no other filters)
CREATE INDEX IF NOT EXISTS idx_items_merchant_created_desc 
ON items(merchant_id, created_at DESC NULLS LAST);

-- Pattern: merchant_id + created_at ASC
CREATE INDEX IF NOT EXISTS idx_items_merchant_created_asc 
ON items(merchant_id, created_at ASC NULLS LAST);

-- Pattern: merchant_id + product_category + created_at DESC
CREATE INDEX IF NOT EXISTS idx_items_merchant_category_created_desc
ON items(merchant_id, product_category, created_at DESC NULLS LAST);

-- Pattern: merchant_id + product_category + created_at ASC
CREATE INDEX IF NOT EXISTS idx_items_merchant_category_created_asc
ON items(merchant_id, product_category, created_at ASC NULLS LAST);

-- 3. TEXT SEARCH INDEX with merchant_id prefix
-- for search
CREATE INDEX IF NOT EXISTS idx_items_merchant_name_trgm 
ON items USING GIN(merchant_id, name gin_trgm_ops);

-- 4. COVERING INDEX for fast lookups without table access
-- When you know the exact item_id
CREATE INDEX IF NOT EXISTS idx_items_id_merchant_covering
ON items(id, merchant_id) INCLUDE (name, product_category, price, image_url, created_at);

-- 5. PARTIAL INDEX for active/recent items
-- CREATE INDEX IF NOT EXISTS idx_items_merchant_recent
-- ON items(merchant_id, created_at DESC)
-- WHERE created_at > NOW() - INTERVAL '30 days';