package merchant

import (
	"context"
	"fmt"

	"belimang/internal/infrastructure/cache"
	"belimang/internal/infrastructure/database"
	logger "belimang/internal/pkg/logging"

	"github.com/google/uuid"
)

// MerchantService handles Merchant business logic
type MerchantService struct {
	cache *cache.RedisCache
	db    database.Querier
}

// NewMerchantService creates a new MerchantService
func NewMerchantService(cache *cache.RedisCache, database database.Querier) *MerchantService {
	return &MerchantService{cache: cache, db: database}
}

func (s *MerchantService) CreateMerchantService(ctx context.Context, adminID uuid.UUID, req PostMerchantRequest) (PostMerchantResponse, error) {
	logger.InfoCtx(ctx, "Create new merchant process", "id: ", adminID, "request: ", req)
	// Should check cache Admin existed?

	// Should check db admin existed?

	// update db merchant
	rows, err := s.db.CreateMerchant(ctx, database.CreateMerchantParams{
		AdminID:          adminID,
		Name:             req.Name,
		MerchantCategory: req.MerchantCategory,
		ImageUrl:         req.ImageURL,
		Lat:              req.Location.Latitude,
		Lng:              req.Location.Longitude,
	})
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to create merchant", "error: ", err)
		return PostMerchantResponse{}, err
	}

	resp := PostMerchantResponse{
		MerchantID: rows.AdminID.String(),
	}

	// set cache merchant
	s.cache.Set(ctx, fmt.Sprintf(cache.MerchantKey, rows.AdminID), resp, cache.MerchantTTL)
	s.cache.Exists(ctx, fmt.Sprintf(cache.MerchantExistsKey, rows.AdminID))

	logger.InfoCtx(ctx, "Merchant created successfully", "resp: ", resp, "rows: ", rows)
	return resp, nil
}
