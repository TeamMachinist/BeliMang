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

-- name: GetNearestMerchant :many
SELECT 
  m.id, 
  m.name, 
  m.merchant_category, 
  m.image_url, 
  ST_Y(m.location::geometry) AS lat, 
  ST_X(m.location::geometry) AS lon, 
  m.created_at, 
  mi.id AS item_id, 
  COALESCE(mi.name, '') AS item_name, 
  COALESCE(mi.product_category, '') AS product_category, 
  COALESCE(mi.price, 0) AS price, 
  COALESCE(mi.image_url, '') AS item_image_url, 
  mi.created_at AS item_created_at, 
  ST_Distance(
    m.location, 
    ST_SetSRID(ST_MakePoint(sqlc.arg('lon'), sqlc.arg('lat')), 4326)
  ) AS distance 
FROM 
  merchants m
  LEFT JOIN items mi ON mi.merchant_id = m.id 
  WHERE (mi.id <> '00000000-0000-0000-0000-000000000000') OR 
  (sqlc.arg('merchant_id')::uuid = '00000000-0000-0000-0000-000000000000'::uuid 
             OR m.id = sqlc.arg('merchant_id')::uuid)
AND m.name ILIKE CONCAT('%', sqlc.arg('name')::text, '%')
AND ((sqlc.narg('merchant_category')::text = '')
  OR m.merchant_category = sqlc.narg('merchant_category')::text)
ORDER BY 
  distance ASC
LIMIT (sqlc.arg('lmt')::int) OFFSET (sqlc.arg('offs')::int);

-- name: CountNearestMerchants :one
SELECT COUNT(DISTINCT m.id)
FROM merchants m
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
    );
