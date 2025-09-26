// internal/service/item_service.go
package items

import (
	"belimang/internal/infrastructure/database"
	"context"
	"errors"
	"fmt"

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
		Productcategory: database.ProductCategory(req.ProductCategory),
		Price:           req.Price,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create item: %w", err)
	}

	return itemID, nil
}
