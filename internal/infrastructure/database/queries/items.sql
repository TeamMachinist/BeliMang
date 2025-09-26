-- name: CreateItem :one
INSERT INTO items (
    merchant_id,
    name,
    product_category,
    price
) VALUES (
    @merchantId::uuid,
    @name::text,
    @productCategory::product_category,
    @price::bigint
)
RETURNING id;

-- name: MerchantExists :one
SELECT EXISTS(SELECT 1 FROM merchants WHERE id = $1);