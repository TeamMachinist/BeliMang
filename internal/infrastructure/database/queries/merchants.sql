-- name: CreateMerchant :one
INSERT INTO merchants (admin_id, name, merchant_category, image_url, lat, lng, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
RETURNING id, admin_id, name, merchant_category, image_url, lat, lng, created_at;