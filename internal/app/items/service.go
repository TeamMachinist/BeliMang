package items

import (
	"belimang/internal/infrastructure/database"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ItemService struct {
	queries *database.Queries
}

func NewItemService(queries *database.Queries) *ItemService {
	return &ItemService{queries: queries}
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

	// Ambil data
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

	return responses, total, nil
}
