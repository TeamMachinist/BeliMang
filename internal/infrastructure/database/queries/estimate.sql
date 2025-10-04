-- name: GetMerchantLatLong :one
SELECT id, lat, lng
FROM merchants
WHERE id = @merchant_id::uuid;

-- name: GetItemPrice :one
SELECT price
FROM items
WHERE id = @item_id::uuid AND merchant_id = @merchant_id::uuid;

-- name: GetMerchantsLatLong :many
SELECT id, lat, lng
FROM merchants
WHERE id = ANY(@merchant_id::uuid[]);

-- name: GetItemPricesByIDsAndMerchants :many
SELECT i.id, i.merchant_id, i.price
FROM items i
JOIN (
    SELECT 
        UNNEST(@item_id::uuid[]) AS item_id,
        UNNEST(@merchant_id::uuid[]) AS merchant_id
) AS pairs ON i.id = pairs.item_id AND i.merchant_id = pairs.merchant_id;

-- name: GetEstimateById :one
SELECT id, user_id, user_lat, user_lng, total_price, estimated_delivery_time_in_minutes, created_at
FROM estimates
WHERE id = $1::uuid;

-- name: CreateEstimate :one
INSERT INTO estimates (
    user_id, user_lat, user_lng, total_price, estimated_delivery_time_in_minutes
) VALUES (
    $1, $2, $3, $4, $5
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

-- name: GetAllMerchantsWithItemsSortedByH3Distance :many
SELECT
    m.id AS merchant_id,
    m.name AS merchant_name,
    m.merchant_category,
    m.image_url AS merchant_image_url,
    m.lat,
    m.lng,
    m.created_at AS merchant_created_at,
    i.id AS item_id,
    i.name AS item_name,
    i.product_category,
    i.price,
    i.image_url AS item_image_url,
    i.created_at AS item_created_at,
    h3_grid_distance(
        h3_latlng_to_cell(Point($1, $2), 10),
        m.h3_index
    ) AS h3_distance
FROM merchants m
JOIN items i ON m.id = i.merchant_id
WHERE ($3 = '' OR m.name ILIKE '%' || $3 || '%')
ORDER BY h3_distance ASC, m.created_at DESC, i.created_at ASC; 
