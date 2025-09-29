-- Enable PostGIS (should already be enabled in this image, but safe to run)
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create merchants table
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(30) NOT NULL CHECK (LENGTH(name) >= 2),
    merchant_category TEXT NOT NULL CHECK (
        merchant_category IN (
            'SmallRestaurant', 'MediumRestaurant', 'LargeRestaurant',
            'MerchandiseRestaurant', 'BoothKiosk', 'ConvenienceStore'
        )
    ),
    image_url TEXT NOT NULL,
    lat FLOAT8 NOT NULL CHECK (lat BETWEEN -90 AND 90),
    lng FLOAT8 NOT NULL CHECK (lng BETWEEN -180 AND 180),
    location GEOGRAPHY(POINT, 4326) GENERATED ALWAYS AS (ST_Point(lng, lat)) STORED,
    created_at TIMESTAMPTZ NOT NUll DEFAULT NOW()
);

-- Create spatial index
CREATE INDEX idx_merchants_location_gist ON merchants USING GIST (location);
