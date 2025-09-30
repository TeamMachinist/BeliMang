-- name: GetMerchantLatLong :one
SELECT id, lat, lng
FROM merchants
WHERE id = @merchant_id::uuid;

-- name: GetItemPrice :one
SELECT price
FROM items
WHERE id = @item_id::uuid AND merchant_id = @merchant_id::uuid;

-- name: GetEstimateById :one
SELECT id, user_lat, user_lng, total_price, estimated_delivery_time_in_minutes, created_at
FROM estimates
WHERE id = $1::uuid;

-- name: CreateEstimate :one
INSERT INTO estimates (
    user_lat, user_lng, total_price, estimated_delivery_time_in_minutes
) VALUES (
    $1, $2, $3, $4
)
RETURNING id, total_price, estimated_delivery_time_in_minutes;

-- name: CreateEstimateOrder :exec
INSERT INTO estimate_orders (
    estimate_id, merchant_id, is_starting_point
) VALUES (
    @estimate_id, @merchant_id, @is_starting_point
);

-- name: CreateEstimateOrderItem :exec
INSERT INTO estimate_order_items (
    estimate_order_id, item_id, quantity
) VALUES (
    @estimate_order_id, @item_id, @quantity
);

-- name: GetEstimateOrderIds :many
SELECT id, merchant_id
FROM estimate_orders
WHERE estimate_id = @estimate_id
ORDER BY id;