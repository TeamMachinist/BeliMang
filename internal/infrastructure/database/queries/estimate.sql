-- name: GetMerchantLatLong :one
SELECT id, lat, lng
FROM merchants
WHERE id = @merchant_id::uuid;

-- name: GetItemPrice :one
SELECT price
FROM items
WHERE id = @item_id::uuid AND merchant_id = @merchant_id::uuid;

-- name: CreateEstimate :one
INSERT INTO estimates (
    user_lat, user_lng, orders, total_price, estimated_delivery_time_in_minutes
) VALUES (
    @user_lat, @user_lng, @orders, @total_price, @estimated_delivery_time_in_minutes
)
RETURNING id, total_price, estimated_delivery_time_in_minutes;