package merchant

import (
	"context"
	"fmt"
	"time"

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
	logger.InfoCtx(ctx, "Create new merchant process", "id", adminID, "request", req)
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
		logger.ErrorCtx(ctx, "Failed to create merchant", "error", err)
		return PostMerchantResponse{}, err
	}

	resp := PostMerchantResponse{
		MerchantID: rows.ID.String(),
	}

	// set cache merchant
	s.cache.Set(ctx, fmt.Sprintf(cache.MerchantKey, rows.ID), resp, cache.MerchantTTL)
	s.cache.Exists(ctx, fmt.Sprintf(cache.MerchantExistsKey, rows.ID))

	logger.InfoCtx(ctx, "Merchant created successfully", "resp", resp, "rows", rows)
	return resp, nil
}

// SearchMerchantsService searches merchants using filter params
func (s *MerchantService) SearchMerchantsService(ctx context.Context, filter MerchantFilter) (GetMerchantsResponse, error) {
	logger.InfoCtx(ctx, "Create search merchants process", "merchantId", filter.MerchantID, "name", filter.Name, "category", filter.MerchantCategory, "sort", filter.CreatedAtSort)

	var merchantId uuid.UUID
	if filter.MerchantID != "" {
		id, err := uuid.Parse(filter.MerchantID)
		if err == nil {
			merchantId = id
		}
	}

	merchantCategory := filter.MerchantCategory
	if merchantCategory != "" {
		if _, ok := validMerchantCategories[merchantCategory]; !ok {
			return GetMerchantsResponse{
				Data: []MerchantItem{},
				Meta: Meta{Limit: filter.Limit, Offset: filter.Offset, Total: 0},
			}, nil
		}
	}

	limit := filter.Limit
	offset := filter.Offset
	if limit <= 0 {
		limit = 5
	}
	if offset < 0 {
		offset = 0
	}

	logger.DebugCtx(ctx, "Query search merchants", "merchantId", merchantId, "name", filter.Name, "category", merchantCategory, "sort", filter.CreatedAtSort)
	if filter.CreatedAtSort == "asc" {
		rows, err := s.db.SearchMerchantsAsc(ctx, database.SearchMerchantsAscParams{
			Column1: merchantId,
			Column2: filter.Name,
			Column3: filter.MerchantCategory,
			Limit:   int32(limit),
			Offset:  int32(offset),
		})
		if err != nil {
			return GetMerchantsResponse{}, err
		}

		// Map DB rows to response
		items := make([]MerchantItem, 0, len(rows))
		for _, r := range rows {
			items = append(items, MerchantItem{
				MerchantID:       r.ID.String(),
				Name:             r.Name,
				MerchantCategory: r.MerchantCategory,
				ImageURL:         r.ImageUrl,
				Location:         Location{Latitude: r.Lat, Longitude: r.Lng},
				CreatedAt:        r.CreatedAt.Format(time.RFC3339Nano),
			})
		}
		logger.InfoCtx(ctx, "Merchant searched successfully", "rows: ", rows)
		// For total, you may want a count query, but for now, use len(rows)
		meta := Meta{Limit: limit, Offset: offset, Total: int(rows[0].TotalCount)}
		return GetMerchantsResponse{Data: items, Meta: meta}, nil
	}

	rows, err := s.db.SearchMerchantsDesc(ctx, database.SearchMerchantsDescParams{
		Column1: merchantId,
		Column2: filter.Name,
		Column3: filter.MerchantCategory,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return GetMerchantsResponse{}, err
	}

	// Map DB rows to response
	items := make([]MerchantItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, MerchantItem{
			MerchantID:       r.ID.String(),
			Name:             r.Name,
			MerchantCategory: r.MerchantCategory,
			ImageURL:         r.ImageUrl,
			Location:         Location{Latitude: r.Lat, Longitude: r.Lng},
			CreatedAt:        r.CreatedAt.Format(time.RFC3339Nano),
		})
	}
	// For total, you may want a count query, but for now, use len(rows)
	meta := Meta{Limit: limit, Offset: offset, Total: int(rows[0].TotalCount)}
	logger.InfoCtx(ctx, "Merchant searched successfully", "rows: ", rows)
	return GetMerchantsResponse{Data: items, Meta: meta}, nil
}
