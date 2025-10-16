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

-- name: GetNearestMerchant :many
WITH user_location AS (
    SELECT 
        ST_SetSRID(ST_Point(sqlc.arg('user_lng')::float, sqlc.arg('user_lat')::float), 4326)::GEOGRAPHY AS point,
        sqlc.arg('user_lng')::float - 0.5 AS min_lng,
        sqlc.arg('user_lat')::float - 0.5 AS min_lat,
        sqlc.arg('user_lng')::float + 0.5 AS max_lng,
        sqlc.arg('user_lat')::float + 0.5 AS max_lat
),
filtered_merchants AS (
    SELECT 
        m.id, 
        m.name,
        m.merchant_category,
        m.image_url,
        m.lat,
        m.lng,
        m.created_at,
        ST_Distance(m.location, ul.point) AS distance_meters
    FROM merchants m
    CROSS JOIN user_location ul
    WHERE
        m.location && ST_MakeEnvelope(ul.min_lng, ul.min_lat, ul.max_lng, ul.max_lat, 4326)
        AND (sqlc.arg('merchant_id')::uuid = '00000000-0000-0000-0000-000000000000'::uuid 
             OR m.id = sqlc.arg('merchant_id')::uuid)
        AND (sqlc.arg('merchant_category')::text = '' 
             OR m.merchant_category = sqlc.arg('merchant_category')::text)
        AND (
            sqlc.arg('search_name')::text = ''
            OR m.name ILIKE '%' || sqlc.arg('search_name')::text || '%'
            OR EXISTS (
                SELECT 1 
                FROM items i 
                WHERE i.merchant_id = m.id 
                  AND i.name ILIKE '%' || sqlc.arg('search_name')::text || '%'
                LIMIT 1
            )
        )
    ORDER BY distance_meters ASC
    LIMIT sqlc.arg('limit_rows') OFFSET sqlc.arg('offset_rows')
)
SELECT
    fm.id AS merchant_id,
    fm.name AS merchant_name,
    fm.merchant_category,
    fm.image_url AS merchant_image_url,
    fm.lat,
    fm.lng,
    fm.created_at AS merchant_created_at,
    fm.distance_meters,
    COALESCE(
        json_agg(
            json_build_object(
                'id', i.id,
                'name', i.name,
                'product_category', i.product_category,
                'price', i.price,
                'image_url', i.image_url,
                'created_at', i.created_at
            ) ORDER BY i.created_at ASC, i.id ASC
        ) FILTER (WHERE i.id IS NOT NULL),
        '[]'::json
    ) AS items
FROM filtered_merchants fm
LEFT JOIN items i ON fm.id = i.merchant_id
GROUP BY 
    fm.id, 
    fm.name, 
    fm.merchant_category, 
    fm.image_url, 
    fm.lat, 
    fm.lng, 
    fm.created_at, 
    fm.distance_meters
ORDER BY fm.distance_meters ASC;

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
