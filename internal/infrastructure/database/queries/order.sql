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


-- name: GetOrdersWithDetails :many
SELECT 
    o.id AS order_id,
    -- o.user_id,
    -- o.estimate_id,
    -- o.total_price,
    -- o.estimated_delivery_time_in_minutes,
    -- o.created_at AS order_created_at,
    om.id AS order_merchant_id,
    -- om.is_starting_point,
    -- om.created_at AS order_merchant_created_at,
    m.id AS merchant_id,
    m.name AS merchant_name,
    m.merchant_category,
    m.image_url AS merchant_image_url,
    m.lat AS merchant_lat,
    m.lng AS merchant_lng,
    m.created_at AS merchant_created_at,
    oi.id AS order_item_id,
    oi.quantity,
    -- oi.created_at AS order_item_created_at,
    i.id AS item_id,
    i.name AS item_name,
    i.product_category,
    i.price AS item_price,
    i.image_url AS item_image_url,
    i.created_at AS item_created_at
FROM orders o
INNER JOIN order_merchants om ON o.id = om.order_id
INNER JOIN merchants m ON om.merchant_id = m.id
INNER JOIN order_items oi ON om.id = oi.order_merchant_id
INNER JOIN items i ON oi.item_id = i.id
WHERE o.user_id = $1
    AND ($2::uuid IS NULL OR $2::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR m.id = $2)
    AND ($3::text IS NULL OR $3 = '' OR m.name ILIKE '%' || $3 || '%' OR i.name ILIKE '%' || $3 || '%')
    AND ($4::text IS NULL OR $4 = '' OR m.merchant_category = $4)
ORDER BY o.created_at DESC, om.created_at ASC, oi.created_at ASC
LIMIT $5 OFFSET $6;