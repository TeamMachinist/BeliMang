CREATE TABLE estimates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_lat FLOAT8 NOT NULL,
    user_lng FLOAT8 NOT NULL,
    orders JSONB NOT NULL,
    total_price BIGINT NOT NULL,
    estimated_delivery_time_in_minutes INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);