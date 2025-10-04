-- name: CreateItem :one
INSERT INTO items (
    merchant_id,
    name,
    product_category,
    price,
    image_url
) VALUES (
    @merchantId::uuid,
    @name::text,
    @productCategory::text,
    @price::bigint,
    @imageUrl::text
)
RETURNING id;

-- name: MerchantExists :one
SELECT EXISTS(SELECT 1 FROM merchants WHERE id = $1);

-- name: ListItemsByMerchant :many
SELECT id, merchant_id, name, product_category, price, image_url, created_at
FROM items
WHERE merchant_id = @merchant_id
    AND (@item_id::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR id = @item_id::uuid)
    AND (@name::text IS NULL OR @name::text = '' OR name ILIKE '%' || @name::text || '%')
    AND (@product_category::text IS NULL OR @product_category::text = '' OR product_category = @product_category)
ORDER BY
    CASE WHEN @created_at_order = 'asc' THEN created_at END ASC NULLS LAST,
    CASE WHEN @created_at_order = 'desc' THEN created_at END DESC NULLS LAST,
    created_at DESC
LIMIT @limitPage OFFSET @offsetPage;

-- name: CountItemsByMerchant :one
SELECT COUNT(*)
FROM items
WHERE merchant_id = @merchant_id
  AND (@item_id::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR id = @item_id::uuid)
  AND (@name::text = '' OR name ILIKE '%' || @name::text || '%')
  AND (@product_category::text = '' OR product_category = @product_category::text);