-- name: CreateMerchant :one
INSERT INTO merchants (admin_id, name, merchant_category, image_url, lat, lng, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
RETURNING id, admin_id, name, merchant_category, image_url, lat, lng, created_at;

-- name: SearchMerchantsDesc :many
SELECT 
    id,
    name,
    merchant_category,
    image_url,
    lat,
    lng,
    created_at
FROM merchants
WHERE
    ($1::uuid IS NULL OR $1 = '00000000-0000-0000-0000-000000000000'::uuid OR id = $1)
    AND ($2::text IS NULL OR $2 = '' OR name ILIKE '%' || $2 || '%')
    AND ($3::text IS NULL OR $3 = '' OR merchant_category = $3)
ORDER BY 
    created_at DESC
LIMIT $4
OFFSET $5;

-- name: SearchMerchantsAsc :many
SELECT 
    id,
    name,
    merchant_category,
    image_url,
    lat,
    lng,
    created_at
FROM merchants
WHERE
    ($1::uuid IS NULL OR $1 = '00000000-0000-0000-0000-000000000000'::uuid OR id = $1)
    AND ($2::text IS NULL OR $2 = '' OR name ILIKE '%' || $2 || '%')
    AND ($3::text IS NULL OR $3 = '' OR merchant_category = $3)
ORDER BY 
    created_at ASC
LIMIT $4 OFFSET $5;

-- name: CountSearchMerchants :one
SELECT COUNT(id)
FROM merchants
WHERE
    ($1::uuid IS NULL OR $1 = '00000000-0000-0000-0000-000000000000'::uuid OR id = $1)
    AND ($2::text IS NULL OR $2 = '' OR name ILIKE '%' || $2 || '%')
    AND ($3::text IS NULL OR $3 = '' OR merchant_category = $3);