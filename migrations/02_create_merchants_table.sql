-- Enable PostGIS (should already be enabled in this image, but safe to run)
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS h3;
-- Enable pg_trgm for trigram-based text search (ILIKE optimization)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create merchants table
CREATE TABLE IF NOT EXISTS merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(30) NOT NULL CHECK (LENGTH(name) >= 2 AND LENGTH(name) <= 30),
    merchant_category VARCHAR(30) NOT NULL CHECK (
        merchant_category IN (
            'SmallRestaurant', 'MediumRestaurant', 'LargeRestaurant',
            'MerchandiseRestaurant', 'BoothKiosk', 'ConvenienceStore'
        )
    ),
    image_url TEXT NOT NULL,
    lat FLOAT8 NOT NULL CHECK (lat BETWEEN -90 AND 90),
    lng FLOAT8 NOT NULL CHECK (lng BETWEEN -180 AND 180),
    location GEOGRAPHY(POINT, 4326) GENERATED ALWAYS AS (ST_Point(lng, lat)) STORED,
    h3_cell BIGINT GENERATED ALWAYS AS (h3_geo_to_cell(lat, lng, 10)) STORED,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- Create spatial index
CREATE INDEX IF NOT EXISTS idx_merchants_location_gist ON merchants USING GIST (location);

-- Create trigram-based text search index for name search (supports ILIKE '%...%')
CREATE INDEX IF NOT EXISTS idx_merchants_name_trgm ON merchants USING GIN(name gin_trgm_ops);

-- Create merchant category index
CREATE INDEX IF NOT EXISTS idx_merchants_category ON merchants(merchant_category);

-- For descending order (newest first - most common)
CREATE INDEX IF NOT EXISTS idx_merchants_created_at_desc ON merchants(created_at DESC NULLS LAST);

-- For ascending order (oldest first)
CREATE INDEX IF NOT EXISTS idx_merchants_created_at_asc ON merchants(created_at ASC NULLS LAST);

-- Composite indexes (filter + sort)
CREATE INDEX IF NOT EXISTS idx_merchants_category_created_desc 
ON merchants(merchant_category, created_at DESC NULLS LAST);

CREATE INDEX IF NOT EXISTS idx_merchants_category_created_asc 
ON merchants(merchant_category, created_at ASC NULLS LAST);

-- Covering index for COUNT optimization
CREATE INDEX IF NOT EXISTS idx_merchants_category_id ON merchants(merchant_category, id);
