\c belimang;
-- Create merchants table
CREATE TABLE IF NOT EXISTS merchants (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
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
    h3_index H3INDEX GENERATED ALWAYS AS (h3_latlng_to_cell(POINT(lat,lng), 10)) STORED,
    location GEOGRAPHY(POINT, 4326) GENERATED ALWAYS AS (ST_Point(lng, lat)) STORED,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Create indexes
CREATE INDEX IF NOT EXISTS idx_merchants_location_gist ON merchants USING GIST (location);
CREATE INDEX IF NOT EXISTS idx_merchants_name_trgm ON merchants USING GIN(name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_merchants_category ON merchants(merchant_category);
CREATE INDEX IF NOT EXISTS idx_merchants_created_at_desc ON merchants(created_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_merchants_created_at_asc ON merchants(created_at ASC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_merchants_category_created_desc ON merchants(merchant_category, created_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_merchants_category_created_asc ON merchants(merchant_category, created_at ASC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_merchants_category_id ON merchants(merchant_category, id);
