package items

import (
	"belimang/internal/infrastructure/cache"
	"belimang/internal/infrastructure/database"
	logger "belimang/internal/pkg/logging"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ItemService struct {
	queries *database.Queries
	cache   *cache.RedisCache
}

func NewItemService(queries *database.Queries, cache *cache.RedisCache) *ItemService {
	return &ItemService{
		queries: queries,
		cache:   cache,
	}
}

func (s *ItemService) CreateItem(ctx context.Context, merchantID uuid.UUID, req CreateItemRequest) (uuid.UUID, error) {
	exists, err := s.queries.MerchantExists(ctx, merchantID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to check merchant: %w", err)
	}
	if !exists {
		return uuid.Nil, errors.New("merchant not found")
	}

	itemID, err := s.queries.CreateItem(ctx, database.CreateItemParams{
		Merchantid:      merchantID,
		Name:            req.Name,
		Productcategory: req.ProductCategory,
		Price:           req.Price,
		Imageurl:        req.ImageUrl,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create item: %w", err)
	}

	// Invalidasi cache terkait merchant menggunakan pattern matching
	err = s.invalidateMerchantItemsCache(ctx, merchantID)
	if err != nil {
		logger.WarnCtx(ctx, "Failed to invalidate merchant items cache", "merchantID", merchantID, "error", err)
	}

	return itemID, nil
}

func (s *ItemService) ListItems(ctx context.Context, merchantID uuid.UUID, req ListItemsRequest) ([]ItemResponse, int64, error) {
	exists, err := s.queries.MerchantExists(ctx, merchantID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check merchant: %w", err)
	}
	if !exists {
		return nil, 0, errors.New("merchant not found")
	}

	cacheKey := s.generateListItemsCacheKey(merchantID, req)

	var cachedResult struct {
		Items []ItemResponse
		Total int64
	}

	err = s.cache.Get(ctx, cacheKey, &cachedResult)
	if err == nil {
		logger.DebugCtx(ctx, "ListItems cache hit", "key", cacheKey)
		return cachedResult.Items, cachedResult.Total, nil
	}

	logger.DebugCtx(ctx, "ListItems cache miss, fetching from database", "key", cacheKey)

	var pgCategory string
	if req.ProductCategory != nil {
		cat := *req.ProductCategory
		allowed := map[string]bool{
			"Beverage": true, "Food": true, "Snack": true,
			"Condiments": true, "Additions": true,
		}
		if allowed[cat] {
			pgCategory = cat
		} else {
			pgCategory = ""
		}
	} else {
		pgCategory = ""
	}

	var order interface{}
	if req.CreatedAtOrder != nil {
		val := strings.ToLower(*req.CreatedAtOrder)
		if val == "asc" || val == "desc" {
			order = val
		} else {
			order = nil
		}
	} else {
		order = nil
	}

	var itemID uuid.UUID
	if req.ItemID != nil {
		itemID = *req.ItemID
	} else {
		itemID = uuid.Nil
	}

	var name string
	if req.Name != nil {
		name = *req.Name
	} else {
		name = ""
	}

	items, err := s.queries.ListItemsByMerchant(ctx, database.ListItemsByMerchantParams{
		MerchantID:      merchantID,
		Limitpage:       req.Limit,
		Offsetpage:      req.Offset,
		ItemID:          itemID,
		Name:            name,
		ProductCategory: pgCategory,
		CreatedAtOrder:  order,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list items: %w", err)
	}

	total, err := s.queries.CountItemsByMerchant(ctx, database.CountItemsByMerchantParams{
		MerchantID:      merchantID,
		ItemID:          itemID,
		Name:            name,
		ProductCategory: pgCategory,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count items: %w", err)
	}

	responses := make([]ItemResponse, len(items))
	for i, item := range items {
		responses[i] = ItemResponse{
			ItemID:          item.ID.String(),
			Name:            item.Name,
			ProductCategory: string(item.ProductCategory),
			Price:           item.Price,
			ImageUrl:        item.ImageUrl,
			CreatedAt:       item.CreatedAt.Format(time.RFC3339Nano),
		}
	}

	cacheData := struct {
		Items []ItemResponse
		Total int64
	}{
		Items: responses,
		Total: total,
	}

	err = s.cache.Set(ctx, cacheKey, cacheData, cache.ProductListTTL)
	if err != nil {
		logger.WarnCtx(ctx, "Failed to cache ListItems result", "key", cacheKey, "error", err)
	}

	return responses, total, nil
}

// generateListItemsCacheKey generates a consistent cache key based on the request parameters
func (s *ItemService) generateListItemsCacheKey(merchantID uuid.UUID, req ListItemsRequest) string {
	keyParts := []string{
		"items",
		"list",
		merchantID.String(),
		fmt.Sprintf("limit:%d", req.Limit),
		fmt.Sprintf("offset:%d", req.Offset),
	}

	if req.ItemID != nil {
		keyParts = append(keyParts, fmt.Sprintf("item:%s", req.ItemID.String()))
	}

	if req.Name != nil {
		keyParts = append(keyParts, fmt.Sprintf("name:%s", *req.Name))
	}

	if req.ProductCategory != nil {
		keyParts = append(keyParts, fmt.Sprintf("category:%s", *req.ProductCategory))
	}

	if req.CreatedAtOrder != nil {
		keyParts = append(keyParts, fmt.Sprintf("order:%s", strings.ToLower(*req.CreatedAtOrder)))
	}

	return strings.Join(keyParts, ":")
}

// invalidateMerchantItemsCache menghapus semua cache terkait item milik merchant
func (s *ItemService) invalidateMerchantItemsCache(ctx context.Context, merchantID uuid.UUID) error {
	pattern := fmt.Sprintf("items:*:%s*", merchantID.String())

	// Gunakan SCAN untuk mencari dan menghapus key secara aman
	err := s.scanAndDeleteKeys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to scan and delete keys with pattern %s: %w", pattern, err)
	}

	logger.DebugCtx(ctx, "Invalidated merchant items cache", "merchantID", merchantID, "pattern", pattern)
	return nil
}

// scanAndDeleteKeys menggunakan SCAN untuk mencari dan menghapus key sesuai pattern
func (s *ItemService) scanAndDeleteKeys(ctx context.Context, pattern string) error {
	iter := s.cache.Client().Scan(ctx, 0, pattern, 0).Iterator()

	var keysToDelete []string
	for iter.Next(ctx) {
		keysToDelete = append(keysToDelete, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("scan error: %w", err)
	}

	if len(keysToDelete) > 0 {
		err := s.cache.Client().Del(ctx, keysToDelete...).Err()
		if err != nil {
			return fmt.Errorf("delete error: %w", err)
		}

		logger.DebugCtx(ctx, "Deleted keys", "pattern", pattern, "count", len(keysToDelete))
	}

	return nil
}

// InvalidateItemCache menghapus cache item tertentu
func (s *ItemService) InvalidateItemCache(ctx context.Context, itemID uuid.UUID) error {
	cacheKey := fmt.Sprintf("item:%s", itemID.String())
	return s.cache.Delete(ctx, cacheKey)
}

// InvalidateItemsByCategory menghapus cache untuk semua item dalam kategori tertentu
func (s *ItemService) InvalidateItemsByCategory(ctx context.Context, category string) error {
	pattern := fmt.Sprintf("items:*:category:%s*", category)
	return s.scanAndDeleteKeys(ctx, pattern)
}

// InvalidateItemsByName menghapus cache untuk item dengan nama tertentu
func (s *ItemService) InvalidateItemsByName(ctx context.Context, name string) error {
	pattern := fmt.Sprintf("items:*:name:%s*", name)
	return s.scanAndDeleteKeys(ctx, pattern)
}
