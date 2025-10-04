-- name: CreateOrderFromEstimate :one
INSERT INTO orders (
    user_id, estimate_id, total_price, estimated_delivery_time_in_minutes
) 
SELECT 
    user_id, id, total_price, estimated_delivery_time_in_minutes
FROM estimates
WHERE id = $1::uuid
RETURNING id, total_price, estimated_delivery_time_in_minutes;

-- name: GetEstimateOrderDetails :many
SELECT 
    eo.merchant_id,
    eo.is_starting_point,
    eoi.item_id,
    eoi.quantity
FROM estimate_orders eo
JOIN estimate_order_items eoi ON eo.id = eoi.estimate_order_id
WHERE eo.estimate_id = $1::uuid
ORDER BY eo.id;

-- name: CreateOrderMerchant :one
INSERT INTO order_merchants (
    order_id, merchant_id, is_starting_point
)
VALUES ($1, $2, $3)
RETURNING id;

-- name: CreateOrderItem :exec
INSERT INTO order_items (
    order_merchant_id, item_id, quantity
)
VALUES ($1, $2, $3);

-- name: GetOrderById :one
SELECT id, estimate_id, total_price, estimated_delivery_time_in_minutes, created_at
FROM orders
WHERE id = $1::uuid;