-- name: GetMerchantLatLong :one
SELECT id, lat, lng FROM merchants WHERE id = @merchant_id;

-- name: GetItemPrice :one
SELECT price FROM items WHERE id = @item_id AND merchant_id = @merchant_id;

-- name: ValidateMerchantsWithin3km :many
-- Gunakan PostGIS ST_Distance (akurat, berbasis elipsoid)
SELECT id
FROM merchants
WHERE id = ANY(@merchant_ids::uuid[])
  AND ST_Distance(location, ST_Point(@user_lng, @user_lat)::geography) > 3000;

-- name: CreateEstimate :one
INSERT INTO estimates (
    user_location,
    orders,
    total_price,
    estimated_delivery_time_in_minutes
) VALUES (
    ST_SetSRID(ST_MakePoint(@user_lng, @user_lat), 4326)::geography,
    @orders,
    @total_price,
    @estimated_delivery_time_in_minutes
)
RETURNING id, total_price, estimated_delivery_time_in_minutes;