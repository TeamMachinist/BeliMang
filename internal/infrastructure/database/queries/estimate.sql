
-- name: GetMerchantLatLong :one
SELECT id, lat, lng, h3_index::h3index
FROM merchants
WHERE id = @merchant_id::uuid;

-- name: GetItemPrice :one
SELECT price
FROM items
WHERE id = @item_id::uuid AND merchant_id = @merchant_id::uuid;

-- name: GetMerchantsLatLong :many
SELECT id, lat, lng,h3_index::h3index
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

-- name: GetNearestMerchant :many
WITH user_location AS (
    SELECT ST_SetSRID(ST_Point(@user_lng, @user_lat), 4326)::GEOGRAPHY AS point
),
filtered_merchants AS (
    SELECT DISTINCT m.id, ST_Distance(m.location, ul.point) AS distance_meters
    FROM merchants m
    CROSS JOIN user_location ul
    LEFT JOIN items i ON m.id = i.merchant_id
    WHERE
        (@merchant_id::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR m.id = @merchant_id)
        AND (
            @search_name::text = ''
            OR m.name ILIKE '%' || @search_name || '%'
            OR EXISTS (
                SELECT 1 FROM items i2 
                WHERE i2.merchant_id = m.id 
                AND i2.name ILIKE '%' || @search_name || '%'
            )
        )
        AND (
            @merchant_category::text = ''
            OR m.merchant_category = @merchant_category
        )
    ORDER BY distance_meters ASC, m.id ASC
    LIMIT @limit_rows OFFSET @offset_rows
)
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
    CAST((POWER(m.lat - ($1), 2) + POWER(m.lng - $2, 2)) AS BIGINT) AS distance_squared 
FROM merchants m
JOIN items i ON m.id = i.merchant_id
WHERE ($3 = '' OR m.name ILIKE '%' || $3 || '%')
ORDER BY distance_squared ASC, m.created_at DESC, i.created_at ASC; 
